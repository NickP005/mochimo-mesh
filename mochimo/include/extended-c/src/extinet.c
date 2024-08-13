/**
 * @private
 * @headerfile extinet.h <extinet.h>
 * @copyright Adequate Systems LLC, 2018-2022. All Rights Reserved.
 * <br />For license information, please refer to ../LICENSE.md
*/

/* include guard */
#ifndef EXTENDED_INTERNET_C
#define EXTENDED_INTERNET_C


#include "extinet.h"
#include "exttime.h"
#include <stdio.h>
#include <string.h>

#if OS_WINDOWS
   #pragma comment(lib, "winhttp.lib")
   #include <winhttp.h>
   #include <wchar.h>

   /* initialize global winsock version 2.2 */
   WSADATA Sockdata;
   WORD Sockverreq = 0x0202;

/* end Windows */
#elif OS_UNIX
   #include <fcntl.h>   /* for fcntl() */
   #include <stdlib.h>  /* for system() */

/* end Unix */
#endif

/* initialize global Sockinuse */
volatile int Sockinuse;

/**
 * Perform a http(s) GET request to a file, fname.
 * @param url Pointer to character array of web URL to download
 * @param fname Pointer to character array of filename to save download as
 * @param timeout Time, in seconds, to wait for download
 * @returns 0 on success, else error code.
*/
int http_get(char *url, char *fname, int timeout)
{
#if OS_WINDOWS
   static const char *default_fname = "index.html";
   wchar_t wcdomain[WEBDOMAIN_MAX], wcpath[WEBPATH_MAX];
   HINTERNET hSession, hConnect, hRequest;
   DWORD bytesAvailable, bytesRead;
   BOOL bResults;
   size_t len;
   FILE *fp;

   /* INTERNET_DEFAULT_PORT evaluates to 80 for HTTP and 443 for HTTPS */
   unsigned int port = INTERNET_DEFAULT_PORT;
   char proto[32] = "https";
   char domain[WEBDOMAIN_MAX] = "\0";
   char path[WEBPATH_MAX] = "/";
   char buf[BUFSIZ];
   char *cp;

   /* url is required... obviously */
   if (url == NULL) return 8;

   /* Standard protocols and ports are automatically assigned, if omitted.
   * URL format: [<protocol>://]<domain>[:<port>]/<path/to/file>
   * example url component extraction (in no real order)...
   *     e.g. https://example-domain.tld:443/path/to/file.ext
   * proto = "https";             port = 443;
   *        domain = "example-domain.tld";  fname = "file.ext";
   *                                path = "/path/to/file.ext"; */

   /* extract protocol (if specified), domain(:port) and path from url */
   if (strstr(url, "://")) {
      sscanf(url, "%15[^:/]://%63[^/]%255s", proto, domain, path);
   } else sscanf(url, "%63[^/]%256s", domain, path);
   /* verify necessary components */
   if (!domain[0] || !path[0]) return 3;
   /* check for, and isolate port from domain, if present */
   cp = strchr(domain, ':');
   if (cp != NULL && cp < &domain[strlen(domain) - 1]) {
      cp = strtok(domain, ":");
      port = atoi(&domain[strlen(cp) + 1]);
   }
   /* if no filename was provided... */
   if (fname == NULL) {
      /* ... extract from path or use static default */
      cp = strrchr(path, '/');
      if (cp != NULL && cp < &path[strlen(path) - 1]) {
         fname = cp + 1;
      } else fname = (char *) default_fname;
   }
   /* convert domain and path to wchar string */
   mbstowcs(wcdomain, domain, WEBDOMAIN_MAX);
   mbstowcs(wcpath, path, WEBPATH_MAX);
   /* open file for writing */
   fp = fopen(fname, "wb");
   if (fp == NULL) return 4;
   /* init internet handles */
   bResults = FALSE;
   hSession = hConnect = hRequest = NULL;
   /* open http session */
   hSession = WinHttpOpen(L"Adequate Systems, LLC. Software",
      WINHTTP_ACCESS_TYPE_DEFAULT_PROXY, WINHTTP_NO_PROXY_NAME,
      WINHTTP_NO_PROXY_BYPASS, 0);
   /* open http connection -- setting timeout parameter */
   if (hSession) {
      timeout = timeout * 1000;  /* convert timeout to milliseconds */
      WinHttpSetTimeouts(hSession, timeout, timeout, timeout, timeout);
      hConnect = WinHttpConnect(hSession, wcdomain, port, 0);
   }
   /* open http request */
   if (hConnect) {
      hRequest = WinHttpOpenRequest(hConnect, L"GET", wcpath, NULL,
         WINHTTP_NO_REFERER, WINHTTP_DEFAULT_ACCEPT_TYPES,
         strncmp(proto, "https", sizeof(proto)) ? WINHTTP_FLAG_SECURE : 0);
   }
   /* send request */
   if (hRequest) {
      bResults = WinHttpSendRequest(hRequest, WINHTTP_NO_ADDITIONAL_HEADERS,
         0, WINHTTP_NO_REQUEST_DATA, 0, 0, 0);
      if (bResults) bResults = WinHttpReceiveResponse(hRequest, NULL);
   }
   /* retrieve data */
   if (bResults) {
      do {
         bytesAvailable = 0;
         WinHttpQueryDataAvailable(hRequest, &bytesAvailable);
         /* limit bytes to size of buf */
         if(bytesAvailable > BUFSIZ) bytesAvailable = BUFSIZ;
         /* read data into buf and write buf to file */
         if(WinHttpReadData(hRequest, buf, bytesAvailable, &bytesRead)) {
            if(fwrite(buf, (size_t) bytesRead, 1, fp) != 1) break;
         }
      } while (bytesAvailable > 0);
   }

   /* cleanup */
   if (hRequest) WinHttpCloseHandle(hRequest);
   if (hConnect) WinHttpCloseHandle(hConnect);
   if (hSession) WinHttpCloseHandle(hSession);
   fclose(fp);

   /* delete partial file downloads */
   if (bytesAvailable > 0) {
      remove(fname);
      return 1;
   }

   return 0;

/* end Windows */
#else
   char cmd[640];

   /* If you've got a solution for a POSIX compliant routine for
   * downloading resources from "secure" (https) locations,
   * without using libraries, take the stage... */

   /* url is required... obviously */
   if (url == NULL) return 8;

   sprintf(cmd,  /* try cURL, with wGET as fallback */
      "curl -sLm %d %.2s %.128s %.128s 2>/dev/null || "
      "wget -qT %d -t 1 %.2s %.128s %.128s 2>/dev/null",
      timeout, fname ? "-o" : "-O", fname ? fname : "", url,
      timeout, fname ? "-O" : "", fname ? fname : "", url);

   return system(cmd);

/* end Unix */
#endif
}  /* end http_get() */

/**
 * Convert IPv4 from 32-bit binary form to numbers-and-dots notation.
 * Place in `char *a` if not `NULL`, else use static char *cp.
 * @param n Pointer to 32-bit value to convert
 * @param a Pointer to character array to place conversion result
 * @returns Character pointer to converted result.
*/
char *ntoa(void *n, char *a)
{
   static char cp[16];
   unsigned char *bp;

   bp = (unsigned char *) n;
   if (a == NULL) a = cp;
   sprintf(a, "%d.%d.%d.%d", bp[0], bp[1], bp[2], bp[3]);

   return a;
}  /* end ntoa() */

/**
 * Convert a network ip address to a 32-bit binary form value by
 * performing an ip address lookup on a network address string.
 * Can use a hostname or numbers-and-dots notation IPv4.
 * @param a Pointer to network ip address character array to convert
 * @returns An IPv4 in 32-bit binary form on success, or 0 on error.
*/
unsigned long aton(char *a)
{
   struct hostent *host;
   struct sockaddr_in addr;

   if(a == NULL) return 0;
   memset(&addr, 0, sizeof(addr));

   host = gethostbyname(a);
   if(host == NULL) return 0;
   memcpy(&(addr.sin_addr.s_addr), host->h_addr_list[0], host->h_length);

   return addr.sin_addr.s_addr;
}  /* end aton() */

/**
 * Obtain an ip address from a valid SOCKET descriptor.
 * @param sd Socket descriptor to obtain ip value from
 * @returns A network ip on success, or INVALID_SOCKET on error.
 * Obtain underlying error code with `sock_errno`.
*/
unsigned long get_sock_ip(SOCKET sd)
{
   static socklen_t addrlen = (socklen_t) sizeof(struct sockaddr_in);
   struct sockaddr_in addr;

   if(getpeername(sd, (struct sockaddr *) &addr, &addrlen) == 0) {
      return addr.sin_addr.s_addr;  /* the unsigned long ip */
   }

   return (unsigned long) SOCKET_ERROR;
}  /* end get_sock_ip() */

/**
 * Perform Socket Startup enabling socket functionality.
 * @retval SOCKET_ERROR if Sockverreq was not met (Windows-only)
 * @retval sock_errno for other socket errors (Windows-only)
 * @retval 0 on success
*/
int sock_startup(void)
{
#if OS_WINDOWS
   /* perform startup request */
   int ecode = WSAStartup(Sockverreq, &Sockdata);
   if (ecode) return ecode;

#endif

   Sockinuse++;  /* increment to enable sockets support */

#if OS_WINDOWS
   /* ensure requested version was obtained */
   if (Sockdata.wVersion != Sockverreq) {
      return -1;
   }

#endif

   return 0;
}  /* end sock_startup() */

/**
 * Obtain the internal state of "sockets-in-use", as modified by functions
 * sock_startup() and sock_cleanup().
 * @returns Zero when sockets have not been started, else non-zero.
*/
int sock_state(void)
{
   return Sockinuse;
}

/**
 * Perform Socket Cleanup, disabling socket functionality.
 * @returns Always returns 0.
 * @note Associated socket support functions will gracefully shutdown.
*/
int sock_cleanup(void)
{
   while(Sockinuse) {

#if OS_WINDOWS
      WSACleanup();

#endif

      Sockinuse--;
   }

   return Sockinuse;
}  /* end sock_cleanup() */

/**
 * Set socket descriptor to non-blocking I/O.
 * @param sd Socket descriptor to set non-blocking I/O
 * @returns 0 on success, or SOCKET_ERROR on error.
 * Obtain underlying error code with `sock_errno`.
*/
int sock_set_nonblock(SOCKET sd)
{
#if OS_WINDOWS
   u_long arg = 1L;
   return ioctlsocket(sd, FIONBIO, &arg);

#elif OS_UNIX
   return fcntl(sd, F_SETFL, fcntl(sd, F_GETFL, 0) | O_NONBLOCK);

#else
   #warning Unexpected OS configuration; sock_set_nonblock() affected.
   return SOCKET_ERROR;

#endif
}  /* end sock_set_nonblock() */

/**
 * Set socket descriptor to blocking I/O.
 * @param sd Socket descriptor to set blocking I/O
 * @returns 0 on success, or SOCKET_ERROR on error.
 * Obtain underlying error code with `sock_errno`.
*/
int sock_set_blocking(SOCKET sd)
{
#if OS_WINDOWS
   u_long arg = 0L;
   return ioctlsocket(sd, FIONBIO, &arg);

#elif OS_UNIX
   return fcntl(sd, F_SETFL, fcntl(sd, F_GETFL, 0) & (~O_NONBLOCK));

#else
   #warning Unexpected OS configuration; sock_set_blocking() affected.
   return SOCKET_ERROR;

#endif
}  /* end sock_set_blocking() */

/**
 * Close an open socket. Does NOT clear socket descriptor value.
 * @param sd socket descriptor to close
*/
int sock_close(SOCKET sd)
{
#if OS_WINDOWS
   return closesocket(sd);

#elif OS_UNIX
   return close(sd);

#else
   #warning Unexpected OS configuration; sock_close() affected.
   return sd;

#endif
}  /* end sock_close() */

/**
 * Initialize an outgoing connection with a network IPv4 address.
 * @param ip 32-bit binary form IPv4 value
 * @param port 16-bit binary form port value
 * @param timeout time, in seconds, to wait for connection
 * @returns Non-blocking SOCKET on success, or INVALID_SOCKET on error.
 * @note sock_startup() MUST be called to enable socket support.
 * @see sock_connect_addr()
*/
SOCKET sock_connect_ip(unsigned long ip, unsigned short port, double timeout)
{
   static socklen_t addrlen = (socklen_t) sizeof(struct sockaddr_in);
   struct sockaddr_in addr;
   time_t start;
   SOCKET sd;
   int ecode;

   if (!Sockinuse) return INVALID_SOCKET;

   sd = socket(AF_INET, SOCK_STREAM, 0);  /* AF_INET = IPv4 */
   if(sd == INVALID_SOCKET) return INVALID_SOCKET;

   memset((char *) &addr, 0, sizeof(addr));
   addr.sin_addr.s_addr = ip;
   addr.sin_family = AF_INET;  /* AF_UNIX */
   /* Convert short integer to network byte order */
   addr.sin_port = htons(port);

   sock_set_nonblock(sd);
   time(&start);

   while (connect(sd, (struct sockaddr *) &addr, addrlen)) {
      ecode = sock_errno;
      if (sock_err_is_success(ecode)) break;
      if (sock_err_is_waiting(ecode) && Sockinuse &&
         (difftime(time(NULL), start) < timeout)) {
         millisleep(1);  /* socket is waiting patiently */
         continue;
      }  /* ... connection is deemed a failure, cleanup... */
      sock_close(sd);
      return INVALID_SOCKET;
   }

   return sd;
}  /* end sock_connect_ip() */

/**
 * Initialize an outgoing connection with a network host address.
 * @param addr Pointer to character array of network host address
 * @param port 16-bit binary form port value
 * @param timeout time, in seconds, to wait for connection
 * @returns Non-blocking SOCKET on success, or INVALID_SOCKET on error.
 * @note sock_startup() MUST be called to enable socket support.
 * @see sock_connect_ip()
*/
SOCKET sock_connect_addr(char *addr, unsigned short port, double timeout)
{
   unsigned long ip = aton(addr);

   return sock_connect_ip(ip, port, timeout);
}  /* end sock_connect_addr() */

/**
 * Receive a packet of data from a socket descriptor into `pkt[len]`.
 * Timeout is ignored if socket is set to blocking.
 * @param sd Socket descriptor to receive packet from
 * @param pkt Pointer to place received data
 * @param len Byte length of data to receive
 * @param flags Flags to pass to recv()
 * @param timeout time, in seconds, to wait for packet
 * @retval (0) for success
 * @retval (1) for end communication
 * @retval (-1) for timeout
*/
int sock_recv(SOCKET sd, void *pkt, int len, int flags, double timeout)
{
   int n, count;
   time_t start;

   time(&start);
   for(n = 0; n < len; ) {
      /* check socket for pending data */
      count = (int) recv(sd, ((char *) pkt) + n, len - n, flags);
      if (count == 0 || !Sockinuse) return 1;  /* end communication */
      if (count > 0) n += count;  /* count recv'd bytes */
      else if(difftime(time(NULL), start) >= timeout) return (-1);
      else millisleep(1);  /* socket is waiting patiently */
   }

   return 0;  /* recv'd packet */
}  /* end sock_recv() */

/**
 * Send a packet of data on SOCKET sd from pkt[len].
 * Timeout is ignored if socket is set to blocking.
 * @param sd Socket descriptor to send packet on
 * @param pkt Pointer to data to send
 * @param len Byte length of data to send
 * @param flags Flags to pass to send()
 * @param timeout time, in seconds, to wait for send
 * @retval (0) for success
 * @retval (1) for end communication
 * @retval (-1) for timeout
*/
int sock_send(SOCKET sd, void *pkt, int len, int flags, double timeout)
{
   int n, count;
   time_t start;

   time(&start);
   for(n = 0; n < len; ) {
      /* check socket for pending data */
      count = (int) send(sd, ((char *) pkt) + n, len - n, flags);
      if (count == 0 || !Sockinuse) return 1;  /* end communication */
      if (count > 0) n += count;  /* count sent bytes */
      else if(difftime(time(NULL), start) >= timeout) return (-1);
      else millisleep(1);  /* socket is waiting patiently */
   }

   return 0;  /* sent packet */
}  /* end sock_send() */

/**
 * Get the IPv4 address of the host.
 * @returns 0 on succesful operation, else non-zero on error.
 * @note Requires sock_startup() to be previously called.
*/
int gethostip(char *name, int namelen)
{
   SOCKET sd;
   struct sockaddr_in addrname;
   socklen_t addrnamelen = sizeof(addrname);

   /* connect to cloudflare dns server 1.1.1.1:53 */
   sd = sock_connect_addr("1.1.1.1", 53, 3);
   if (sd < 0) return SOCKET_ERROR;

   /* get socket name for interpretation */
   if (getsockname(sd, (struct sockaddr *) &addrname, &addrnamelen) == 0) {
      if (inet_ntop(AF_INET, &addrname.sin_addr, name, namelen) != NULL) {
         return sock_close(sd);
      }
   }

   sock_close(sd);
   return SOCKET_ERROR;
}  /* end phostinfo() */

/* end include guard */
#endif
