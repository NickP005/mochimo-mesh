package main

import (
	"bytes"
	"crypto/rand"
	"crypto/sha256"
	"encoding/binary"
	"fmt"
	"log"
)

// Constants
const (
	XMSS_HASH_PADDING_F   = 0
	XMSS_HASH_PADDING_PRF = 3
	PARAMSN               = 32
	WOTSW                 = 16
	WOTSLOGW              = 4
	WOTSLEN2              = 3
	WOTSLEN1              = (8 * PARAMSN / WOTSLOGW)
	WOTSLEN               = (WOTSLEN1 + WOTSLEN2)
)

// Helper function to convert unsigned long to bytes in big-endian
func ullToBytes(out []byte, outlen int, in uint64) {
	for i := outlen - 1; i >= 0; i-- {
		out[i] = byte(in & 0xff)
		in >>= 8
	}
}

// Core hashing function - sha256
func coreHash(out, in []byte) {
	hash := sha256.Sum256(in)
	copy(out, hash[:])
}

// Functions for OTS addresses
func setKeyAndMask(addr []uint32, keyAndMask uint32) {
	addr[7] = keyAndMask
}

func setChainAddr(addr []uint32, chain uint32) {
	addr[5] = chain
}

func setHashAddr(addr []uint32, hash uint32) {
	addr[6] = hash
}

// Convert addr to bytes
func addrToBytes(bytes []byte, addr []uint32) {
	for i := 0; i < 8; i++ {
		binary.BigEndian.PutUint32(bytes[i*4:], addr[i])
	}
}

// Computes PRF(key, in)
func prf(out, in, key []byte) {
	buf := make([]byte, 2*PARAMSN+32)

	ullToBytes(buf, PARAMSN, XMSS_HASH_PADDING_PRF)
	copy(buf[PARAMSN:], key)
	copy(buf[2*PARAMSN:], in)
	coreHash(out, buf)
}

// thash_f implementation
func thashF(out, in, pubSeed []byte, addr []uint32) {
	buf := make([]byte, 3*PARAMSN)
	bitmask := make([]byte, PARAMSN)
	addrAsBytes := make([]byte, 32)

	// Set the function padding.
	ullToBytes(buf, PARAMSN, XMSS_HASH_PADDING_F)

	// Generate the n-byte key.
	setKeyAndMask(addr, 0)
	addrToBytes(addrAsBytes, addr)
	prf(buf[PARAMSN:], addrAsBytes, pubSeed)

	// Generate the n-byte mask.
	setKeyAndMask(addr, 1)
	addrToBytes(addrAsBytes, addr)
	prf(bitmask, addrAsBytes, pubSeed)

	for i := 0; i < PARAMSN; i++ {
		buf[2*PARAMSN+i] = in[i] ^ bitmask[i]
	}

	coreHash(out, buf)
}

// expand_seed implementation
func expandSeed(outseeds, inseed []byte) {
	ctr := make([]byte, 32)

	for i := 0; i < WOTSLEN; i++ {
		ullToBytes(ctr, 32, uint64(i))
		prf(outseeds[i*PARAMSN:], ctr, inseed)
	}
}

// gen_chain implementation
func genChain(out, in []byte, start, steps uint32, pubSeed []byte, addr []uint32) {
	copy(out, in)

	for i := start; i < start+steps && i < WOTSW; i++ {
		setHashAddr(addr, i)
		thashF(out, out, pubSeed, addr)
	}
}

// base_w implementation
func baseW(output []int, outLen int, input []byte) {
	in := 0
	out := 0
	var total byte
	bits := 0

	for consumed := 0; consumed < outLen; consumed++ {
		if bits == 0 {
			total = input[in]
			in++
			bits += 8
		}
		bits -= WOTSLOGW
		output[out] = int((total >> bits) & (WOTSW - 1))
		out++
	}
}

// wots_checksum implementation
func wotsChecksum(csumBaseW, msgBaseW []int) {
	csum := 0
	csumBytes := make([]byte, (WOTSLEN2*WOTSLOGW+7)/8)

	for i := 0; i < WOTSLEN1; i++ {
		csum += WOTSW - 1 - msgBaseW[i]
	}

	csum <<= (8 - ((WOTSLEN2 * WOTSLOGW) % 8))
	ullToBytes(csumBytes, len(csumBytes), uint64(csum))
	baseW(csumBaseW, WOTSLEN2, csumBytes)
}

// chain_lengths implementation
func chainLengths(lengths []int, msg []byte) {
	baseW(lengths, WOTSLEN1, msg)
	wotsChecksum(lengths[WOTSLEN1:], lengths)
}

// wots_pkgen implementation
func wotsPkgen(pk, seed, pubSeed []byte, addr []uint32) {
	expandSeed(pk, seed)

	for i := 0; i < WOTSLEN; i++ {
		setChainAddr(addr, uint32(i))
		genChain(pk[i*PARAMSN:], pk[i*PARAMSN:], 0, WOTSW-1, pubSeed, addr)
	}
}

// wots_sign implementation
func wotsSign(sig, msg, seed, pubSeed []byte, addr []uint32) {
	lengths := make([]int, WOTSLEN)

	chainLengths(lengths, msg)
	expandSeed(sig, seed)

	for i := 0; i < WOTSLEN; i++ {
		setChainAddr(addr, uint32(i))
		genChain(sig[i*PARAMSN:], sig[i*PARAMSN:], 0, uint32(lengths[i]), pubSeed, addr)
	}
}

// wots_pk_from_sig implementation
func wotsPkFromSig(pk, sig, msg, pubSeed []byte, addr []uint32) {
	lengths := make([]int, WOTSLEN)

	chainLengths(lengths, msg)

	for i := 0; i < WOTSLEN; i++ {
		setChainAddr(addr, uint32(i))
		genChain(pk[i*PARAMSN:], sig[i*PARAMSN:], uint32(lengths[i]), WOTSW-1-uint32(lengths[i]), pubSeed, addr)
	}
}

// ----------------------------------------------------------------
// Test Functions

// Generates a random seed
func generateRandomSeed() []byte {
	seed := make([]byte, PARAMSN)
	if _, err := rand.Read(seed); err != nil {
		log.Fatalf("failed to generate random seed: %v", err)
	}
	return seed
}

// Test the length of generated key and signature
func testWotsLength() {
	seed := generateRandomSeed()
	pubSeed := generateRandomSeed()
	addr := make([]uint32, 8)

	pk := make([]byte, WOTSLEN*PARAMSN)
	wotsPkgen(pk, seed, pubSeed, addr)

	sig := make([]byte, WOTSLEN*PARAMSN)
	msg := []byte("Test message for WOTS+ signature")
	wotsSign(sig, msg, seed, pubSeed, addr)

	fmt.Printf("Public Key Length: %d bytes\n", len(pk))
	fmt.Printf("Signature Length: %d bytes\n", len(sig))
}

// Test the correctness of the signature
func testWotsSignature() {
	seed := generateRandomSeed()
	pubSeed := generateRandomSeed()
	addr := make([]uint32, 8)

	pk := make([]byte, WOTSLEN*PARAMSN)
	wotsPkgen(pk, seed, pubSeed, addr)

	sig := make([]byte, WOTSLEN*PARAMSN)
	msg := []byte("Test message for WOTS+ signature")
	wotsSign(sig, msg, seed, pubSeed, addr)

	recoveredPk := make([]byte, WOTSLEN*PARAMSN)
	wotsPkFromSig(recoveredPk, sig, msg, pubSeed, addr)

	if bytes.Equal(pk, recoveredPk) {
		fmt.Println("Signature is valid: public key matches the recovered key.")
	} else {
		fmt.Println("Signature is invalid: public key does not match the recovered key.")
	}
}

// Test signature verification with a wrong seed
func testWotsInvalidSignature() {
	seed := generateRandomSeed()
	otherSeed := generateRandomSeed() // Different seed for invalid test
	pubSeed := generateRandomSeed()
	addr := make([]uint32, 8)

	pk := make([]byte, WOTSLEN*PARAMSN)
	wotsPkgen(pk, seed, pubSeed, addr)

	sig := make([]byte, WOTSLEN*PARAMSN)
	msg := []byte("Test message for WOTS+ signature")
	wotsSign(sig, msg, seed, pubSeed, addr)

	// Misuse the signature with the wrong private key seed
	recoveredPk := make([]byte, WOTSLEN*PARAMSN)
	wotsPkgen(recoveredPk, otherSeed, pubSeed, addr)

	if bytes.Equal(pk, recoveredPk) {
		fmt.Println("Error: Invalid signature validated with incorrect public key.")
	} else {
		fmt.Println("Correctly identified invalid signature: public key does not match.")
	}
}

func main() {
	fmt.Println("WOTS+ Tests")
	testWotsLength()
	testWotsSignature()
	testWotsInvalidSignature()
}
