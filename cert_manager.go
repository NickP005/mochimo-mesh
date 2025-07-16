package main

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"sync"
	"time"
)

// CertManager gestisce il ricaricamento automatico dei certificati SSL
type CertManager struct {
	certFile     string
	keyFile      string
	server       *http.Server
	handler      http.Handler
	port         int
	mu           sync.RWMutex
	stopChan     chan struct{}
	ctx          context.Context
	cancel       context.CancelFunc
	lastCertTime time.Time
}

// NewCertManager crea un nuovo gestore dei certificati
func NewCertManager(certFile, keyFile string, handler http.Handler, port int) *CertManager {
	ctx, cancel := context.WithCancel(context.Background())
	return &CertManager{
		certFile: certFile,
		keyFile:  keyFile,
		handler:  handler,
		port:     port,
		stopChan: make(chan struct{}),
		ctx:      ctx,
		cancel:   cancel,
	}
}

// Start avvia il server HTTPS con ricaricamento automatico dei certificati
func (cm *CertManager) Start() error {
	// Carica i certificati iniziali
	cert, err := cm.loadCertificate()
	if err != nil {
		return fmt.Errorf("failed to load initial certificate: %v", err)
	}

	// Configura il server HTTPS con certificato dinamico
	cm.server = &http.Server{
		Addr:    ":" + strconv.Itoa(cm.port),
		Handler: cm.handler,
		TLSConfig: &tls.Config{
			GetCertificate: cm.getCertificate,
		},
	}

	// Aggiorna il timestamp del certificato
	cm.updateCertificateTime(cert)

	// Avvia il monitoraggio dei certificati in background
	go cm.watchCertificates()

	mlog(2, "§bCertManager: §2Starting HTTPS server on port %d with automatic certificate reloading", cm.port)

	// Avvia il server HTTPS
	err = cm.server.ListenAndServeTLS("", "")
	if err != nil && err != http.ErrServerClosed {
		return fmt.Errorf("HTTPS server failed: %v", err)
	}

	return nil
}

// Stop ferma il server HTTPS
func (cm *CertManager) Stop() error {
	mlog(2, "§bCertManager: §6Stopping HTTPS server...")

	// Segnala al watcher di fermarsi
	cm.cancel()
	close(cm.stopChan)

	// Ferma il server con timeout
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	return cm.server.Shutdown(ctx)
}

// getCertificate restituisce il certificato corrente (chiamata da TLS)
func (cm *CertManager) getCertificate(*tls.ClientHelloInfo) (*tls.Certificate, error) {
	cert, err := cm.loadCertificate()
	if err != nil {
		mlog(1, "§bCertManager: §4Error loading certificate: %v", err)
		return nil, err
	}
	return cert, nil
}

// loadCertificate carica il certificato dai file
func (cm *CertManager) loadCertificate() (*tls.Certificate, error) {
	cm.mu.RLock()
	defer cm.mu.RUnlock()

	cert, err := tls.LoadX509KeyPair(cm.certFile, cm.keyFile)
	if err != nil {
		return nil, err
	}

	return &cert, nil
}

// updateCertificateTime aggiorna il timestamp dell'ultimo caricamento del certificato
func (cm *CertManager) updateCertificateTime(cert *tls.Certificate) {
	cm.mu.Lock()
	defer cm.mu.Unlock()

	if len(cert.Certificate) > 0 {
		if x509Cert, err := x509.ParseCertificate(cert.Certificate[0]); err == nil {
			cm.lastCertTime = x509Cert.NotAfter
			mlog(3, "§bCertManager: §7Certificate expires at: §e%s", cm.lastCertTime.Format(time.RFC3339))
		}
	}
}

// watchCertificates monitora i certificati e li ricarica se necessario
func (cm *CertManager) watchCertificates() {
	ticker := time.NewTicker(30 * time.Minute) // Controlla ogni 30 minuti
	defer ticker.Stop()

	mlog(3, "§bCertManager: §7Started certificate monitoring (checking every 30 minutes)")

	for {
		select {
		case <-cm.ctx.Done():
			mlog(3, "§bCertManager: §6Certificate monitoring stopped")
			return
		case <-ticker.C:
			cm.checkAndReloadCertificates()
		}
	}
}

// checkAndReloadCertificates controlla se i certificati devono essere ricaricati
func (cm *CertManager) checkAndReloadCertificates() {
	// Carica il certificato per controllare la data di scadenza
	cert, err := cm.loadCertificate()
	if err != nil {
		mlog(2, "§bCertManager: §4Error checking certificate: %v", err)
		return
	}

	// Verifica se il certificato è cambiato o sta per scadere
	if len(cert.Certificate) > 0 {
		if x509Cert, err := x509.ParseCertificate(cert.Certificate[0]); err == nil {
			now := time.Now()

			// Controlla se il certificato scade nei prossimi 7 giorni
			daysUntilExpiry := x509Cert.NotAfter.Sub(now).Hours() / 24

			if daysUntilExpiry <= 7 {
				mlog(2, "§bCertManager: §3Certificate expires in %.1f days (%s), will schedule renewal check",
					daysUntilExpiry, x509Cert.NotAfter.Format(time.RFC3339))

				// Se scade entro 24 ore, programma il controllo di rinnovo
				if daysUntilExpiry <= 1 {
					go cm.scheduleRenewalCheck(x509Cert.NotAfter)
				}
			}

			// Controlla se il certificato è stato rinnovato (data di scadenza diversa)
			cm.mu.RLock()
			lastCertTime := cm.lastCertTime
			cm.mu.RUnlock()

			if !x509Cert.NotAfter.Equal(lastCertTime) {
				mlog(1, "§bCertManager: §2Certificate has been renewed! Old expiry: %s, New expiry: %s",
					lastCertTime.Format(time.RFC3339), x509Cert.NotAfter.Format(time.RFC3339))
				cm.updateCertificateTime(cert)
			}

			mlog(5, "§bCertManager: §7Certificate check: expires in %.1f days (%s)",
				daysUntilExpiry, x509Cert.NotAfter.Format(time.RFC3339))
		}
	}
}

// scheduleRenewalCheck programma un controllo di rinnovo del certificato poco dopo la scadenza
func (cm *CertManager) scheduleRenewalCheck(expiryTime time.Time) {
	now := time.Now()

	// Calcola quando il certificato scadrà
	timeUntilExpiry := expiryTime.Sub(now)

	// Aspetta 1 minuto dopo la scadenza prima di provare a ricaricare
	waitTime := timeUntilExpiry + (1 * time.Minute)

	mlog(2, "§bCertManager: §3Scheduling certificate renewal check in %v (1 minute after expiry at %s)",
		waitTime, expiryTime.Format(time.RFC3339))

	// Se il certificato è già scaduto, inizia subito
	if waitTime <= 0 {
		waitTime = 1 * time.Minute
		mlog(2, "§bCertManager: §3Certificate is already expired, starting renewal attempts immediately")
	}

	// Timer per aspettare fino al momento del controllo
	timer := time.NewTimer(waitTime)
	defer timer.Stop()

	select {
	case <-cm.ctx.Done():
		return
	case <-timer.C:
		mlog(1, "§bCertManager: §2Starting certificate renewal attempts...")
		cm.attemptCertificateRenewal()
	}
}

// attemptCertificateRenewal tenta di ricaricare il certificato rinnovato
func (cm *CertManager) attemptCertificateRenewal() {
	ticker := time.NewTicker(1 * time.Minute) // Riprova ogni minuto
	defer ticker.Stop()

	maxAttempts := 60 // Massimo 60 tentativi (1 ora)
	attempts := 0

	for {
		select {
		case <-cm.ctx.Done():
			return
		case <-ticker.C:
			attempts++

			cert, err := cm.loadCertificate()
			if err != nil {
				mlog(2, "§bCertManager: §4Renewal attempt %d/%d failed: %v", attempts, maxAttempts, err)
			} else if len(cert.Certificate) > 0 {
				if x509Cert, err := x509.ParseCertificate(cert.Certificate[0]); err == nil {
					now := time.Now()

					// Controlla se il certificato è stato rinnovato (non è scaduto)
					if x509Cert.NotAfter.After(now) {
						daysUntilExpiry := x509Cert.NotAfter.Sub(now).Hours() / 24

						// Se il certificato è valido per più di 1 giorno, consideralo rinnovato
						if daysUntilExpiry > 1 {
							mlog(1, "§bCertManager: §2Certificate successfully renewed! New expiry: %s (%.1f days)",
								x509Cert.NotAfter.Format(time.RFC3339), daysUntilExpiry)
							cm.updateCertificateTime(cert)
							return
						} else {
							mlog(3, "§bCertManager: §6Certificate loaded but still expires soon (%.1f days), continuing attempts...", daysUntilExpiry)
						}
					} else {
						mlog(3, "§bCertManager: §6Certificate is still expired, attempt %d/%d", attempts, maxAttempts)
					}
				}
			}

			// Se abbiamo raggiunto il numero massimo di tentativi
			if attempts >= maxAttempts {
				mlog(1, "§bCertManager: §4Maximum renewal attempts reached (%d). Stopping renewal attempts.", maxAttempts)
				return
			}
		}
	}
}

// ForceRenewalCheck forza un controllo immediato del certificato (per testing/debugging)
func (cm *CertManager) ForceRenewalCheck() {
	mlog(2, "§bCertManager: §6Manual certificate renewal check triggered")
	cm.checkAndReloadCertificates()
}

// GetCertificateInfo restituisce informazioni sul certificato corrente
func (cm *CertManager) GetCertificateInfo() (map[string]interface{}, error) {
	cert, err := cm.loadCertificate()
	if err != nil {
		return nil, err
	}

	info := make(map[string]interface{})

	if len(cert.Certificate) > 0 {
		if x509Cert, err := x509.ParseCertificate(cert.Certificate[0]); err == nil {
			now := time.Now()
			daysUntilExpiry := x509Cert.NotAfter.Sub(now).Hours() / 24

			info["subject"] = x509Cert.Subject.CommonName
			info["issuer"] = x509Cert.Issuer.CommonName
			info["not_before"] = x509Cert.NotBefore.Format(time.RFC3339)
			info["not_after"] = x509Cert.NotAfter.Format(time.RFC3339)
			info["days_until_expiry"] = daysUntilExpiry
			info["is_expired"] = x509Cert.NotAfter.Before(now)
			info["dns_names"] = x509Cert.DNSNames
		}
	}

	return info, nil
}

// certStatusHandler restituisce lo stato del certificato SSL
func certStatusHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method == "OPTIONS" {
		return
	}

	if !Globals.EnableHTTPS || Globals.CertManager == nil {
		http.Error(w, "HTTPS not enabled", http.StatusServiceUnavailable)
		return
	}

	info, err := Globals.CertManager.GetCertificateInfo()
	if err != nil {
		mlog(3, "§bcertStatusHandler(): §4Error getting certificate info: %v", err)
		http.Error(w, "Error reading certificate", http.StatusInternalServerError)
		return
	}

	response := map[string]interface{}{
		"status": "success",
		"data":   info,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}
