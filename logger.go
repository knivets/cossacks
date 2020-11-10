package main

import (
	"bufio"
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	b64 "encoding/base64"
	"flag"
	"fmt"
	"golang.org/x/crypto/argon2"
	"io"
	"os"
	"os/signal"
	"syscall"
	"time"
)

const KB = 1024
const MB = 1024 * KB

const DEF_BUFF_SIZE = 128 * KB
const MAX_BUFF_SIZE = 1 * MB

const DEF_READ_SPEED = 100
const MAX_READ_SPEED = 3000

func generateNonce(size int) []byte {
	if size < 12 {
		panic("Nonce should be >= 12")
	}
	nonce := make([]byte, size)
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		panic(err.Error())
	}
	return nonce
}

func encrypt(data []byte, key []byte) []byte {
	block, err := aes.NewCipher(key)
	if err != nil {
		panic(err.Error())
	}

	// make sure to not reuse nonce
	nonce := generateNonce(12)

	aesgcm, err := cipher.NewGCM(block)
	if err != nil {
		panic(err.Error())
	}

	ct := aesgcm.Seal(nil, nonce, data, nil)
	out := append([]byte{}, nonce...)
	return append(out, ct...)
}

func decrypt(data []byte, key []byte) []byte {
	block, err := aes.NewCipher(key)
	if err != nil {
		panic(err.Error())
	}

	aesgcm, err := cipher.NewGCM(block)
	if err != nil {
		panic(err.Error())
	}

	nonce := data[0:12]
	ct := data[12:]

	pt, err := aesgcm.Open(nil, nonce, ct, nil)
	if err != nil {
		panic(err.Error())
	}
	return pt
}

func decryptFile(path string, key []byte) {
	// just a helper function to verify that decryption works
	file, err := os.Open(path)
	if err != nil {
		panic(err)
	}
	scanner := bufio.NewScanner(file)

	for scanner.Scan() {
		text := scanner.Text()
		ct, _ := b64.StdEncoding.DecodeString(text)
		pt := decrypt(ct, key)
		fmt.Println(text)
		fmt.Println(string(pt))
		fmt.Println("---")
	}
}

func stretchKey(key []byte) []byte {
	// !!! we generate static salt for debugging purposes
	salt := make([]byte, 16)
	// here's what you should use in production, however we need
	// to find a way to store salt along with they key
	// salt := generateNonce(16)

	// time=1 recommended in RFC
	var time uint32 = 1
	// memory in IDKEY is specified in kb so the actual value is 64mb
	var memory uint32 = 64 * 1024
	var threads uint8 = 1
	var keyLen uint32 = 16
	// https://godoc.org/golang.org/x/crypto/argon2#IDKey
	hsh := argon2.IDKey(key, salt, time, memory, threads, keyLen)
	return hsh
}

func main() {
	var keyFlag = flag.String("log_key", "", "encryption key")
	var pathFlag = flag.String("file_path", "", "path to file where to store the logs")
	var buffFlag = flag.Int("buffer_size", DEF_BUFF_SIZE, "path to file where to store the logs")
	var flowFlag = flag.Int("flow_speed", DEF_READ_SPEED, "path to file where to store the logs")
	var debugFlag = flag.Bool("debug", false, "enable debug mode which expects --file_path and --log_key to decrypt the contents of the file")
	flag.Parse()
	logKey := *keyFlag
	filePath := *pathFlag
	buffSize := *buffFlag
	debug := *debugFlag

	flowSpeed := *flowFlag
	if flowSpeed > MAX_READ_SPEED {
		flowSpeed = MAX_READ_SPEED
	}

	if len(logKey) > 0 {
		if len(logKey) < 4 {
			panic("Key should be >= 4")
		}
	}
	if len(filePath) == 0 {
		panic("Empty path")
	}

	// only generate when key is present
	key := stretchKey([]byte(logKey))

	if debug {
		decryptFile(filePath, key)
		return
	}

	// handle minimal buffer size?
	outFile, fileErr := os.Create(filePath)
	if fileErr != nil {
		//todo
		panic("something went wrong")
	}
	writer := bufio.NewWriterSize(outFile, buffSize)

	sigIntCh := make(chan os.Signal, 1)
	signal.Notify(sigIntCh, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		<-sigIntCh

		// make sure the data has been committed to disk
		writer.Flush()
		outFile.Sync()
		outFile.Close()
		os.Exit(0)
	}()

	scanner := bufio.NewScanner(os.Stdin)
	startTime := time.Now()
	total := 0

	for scanner.Scan() {
		// we chose to ignore the messages, however to make sure we don't lose
		// any messages move this line inside the if block
		text := scanner.Text()
		if total < flowSpeed {
			out := text
			if len(logKey) >= 4 {
				ct := encrypt([]byte(out), key)
				out = b64.StdEncoding.EncodeToString(ct)
			}
			_, writeErr := writer.WriteString(out + "\n")
			if writeErr != nil {
				panic("something went wrong")
			}
			total += 1
		} else if time.Since(startTime) > time.Second {
			startTime = time.Now()
			total = 0
		}
	}
	if err := scanner.Err(); err != nil {
		fmt.Fprintln(os.Stderr, "reading standard input:", err)
	}
}
