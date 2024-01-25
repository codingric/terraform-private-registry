package main

import (
	"bytes"
	"crypto/sha256"
	"fmt"
	"log"
	"strings"

	"golang.org/x/crypto/openpgp"
	"golang.org/x/crypto/openpgp/armor"
	"golang.org/x/crypto/openpgp/packet"
)

func generateShaSum(contents []byte) []byte {
	h := sha256.New()
	h.Write(contents)
	return h.Sum(nil)
}

func signGpg(contents []byte) ([]byte, error) {
	pk, err := getPrivateKey()
	if err != nil {
		return nil, err
	}
	out := new(bytes.Buffer)
	err = openpgp.DetachSign(out, pk, bytes.NewReader(contents), nil)
	if err != nil {
		return nil, err
	}
	return out.Bytes(), nil
}

func getPrivateKey() (*openpgp.Entity, error) {

	entitylist, err := openpgp.ReadArmoredKeyRing(strings.NewReader(string(privateKey)))
	if err != nil {
		log.Fatal(err)
	}

	entity := entitylist[0]

	if entity.PrivateKey != nil && entity.PrivateKey.Encrypted {
		err := entity.PrivateKey.Decrypt(privatePassword)
		if err != nil {
			fmt.Println("Failed to decrypt key")
		}
	}

	for _, subkey := range entity.Subkeys {
		if subkey.PrivateKey != nil && subkey.PrivateKey.Encrypted {
			err := subkey.PrivateKey.Decrypt(privatePassword)
			if err != nil {
				fmt.Println("Failed to decrypt subkey")
			}
		}
	}
	return entity, nil
}

func publicKeyId() string {
	block, _ := armor.Decode(bytes.NewReader(publicKey))
	reader := packet.NewReader(block.Body)
	pkt, _ := reader.Next()
	key, _ := pkt.(*packet.PublicKey)
	return key.KeyIdString()
}
