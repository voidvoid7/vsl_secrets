package secretmap

import (
	"fmt"
	"strings"
	"testing"
)

func TestSecretMap1(t *testing.T) {
	testInputData := []string{"test1", "test2", "test3", "test4", "test5", "test6", "test7", "test8", "test9", "test10"}
	secretMapHolder := NewSecretMapHolder()
	for _, data := range testInputData {
		key, err := secretMapHolder.Set([]byte(data))
		if err != nil {
			t.Fatal(err)
		}
		decryptedData, err := secretMapHolder.Get(key)
		if err != nil {
			t.Fatal(err)
		}
		decryptedDataStr := string(decryptedData)
		if data != decryptedDataStr {
			fmt.Printf("Expected %v, got %v\n", len(data), len(decryptedDataStr))
			fmt.Printf("fooooo")
			t.Fatalf("Expected %s, got %s", data, decryptedDataStr)
		}
	}

}

func TestSecretMap2(t *testing.T) {
	testInputData := []string{"void foo goo ", "foo", "daddy", "test me daddy", "boo loo", "moo", "im o", "", "loo", "who are you?"}
	secretMapHolder := NewSecretMapHolder()
	for _, data := range testInputData {
		key, err := secretMapHolder.Set([]byte(data))
		if err != nil {
			t.Fatal(err)
		}
		decryptedData, err := secretMapHolder.Get(key)
		if err != nil {
			t.Fatal(err)
		}
		decryptedDataStr := string(decryptedData)
		if data != decryptedDataStr {
			t.Fatalf("Expected %s, got %s", data, decryptedDataStr)
		}
	}

}

// max and min length
func TestSecretMap3(t *testing.T) {
	testString := strings.Repeat("a", 1032)
	testInputData := []string{testString, ""}
	secretMapHolder := NewSecretMapHolder()
	for _, data := range testInputData {
		key, err := secretMapHolder.Set([]byte(data))
		if err != nil {
			t.Fatal(err)
		}
		decryptedData, err := secretMapHolder.Get(key)
		if err != nil {
			t.Fatal(err)
		}
		decryptedDataStr := string(decryptedData)
		if data != decryptedDataStr {
			t.Fatalf("Expected %s, got %s", data, decryptedDataStr)
		}
	}

}

// over max length, attempting set
func TestSecretMap4(t *testing.T) {
	testString1 := strings.Repeat("a", 1033)
	testString2 := strings.Repeat("a", 10033)
	testString3 := strings.Repeat("a", 20000)
	testInputData := []string{testString1, testString2, testString3}
	secretMapHolder := NewSecretMapHolder()
	for _, data := range testInputData {
		_, err := secretMapHolder.Set([]byte(data))
		if err == nil {
			t.Fatal(err)
		}
	}
}

// setting correctly, attempting get with too long key
func TestSecretMap5(t *testing.T) {
	testString1 := strings.Repeat("a", 1032)
	testString2 := strings.Repeat("b", 1032)
	testString3 := strings.Repeat("c", 1032)
	testInputData := []string{testString1, testString2, testString3}
	secretMapHolder := NewSecretMapHolder()
	for _, data := range testInputData {
		key, err := secretMapHolder.Set([]byte(data))
		if err != nil {
			t.Fatal(err)
		}
		_, err = secretMapHolder.Get(key + "a")
		if err == nil {
			t.Fatal(err)
		}
	}
}

// getting non existing secret
func TestSecretMap6(t *testing.T) {
	testString1 := strings.Repeat("a", 1032)
	testString2 := strings.Repeat("b", 1032)
	testString3 := strings.Repeat("c", 1032)
	testString4 := "a"
	testInputData := []string{testString1, testString2, testString3, testString4}
	secretMapHolder := NewSecretMapHolder()
	for _, data := range testInputData {
		key, err := secretMapHolder.Set([]byte(data))
		if err != nil {
			t.Fatal(err)
		}
		_, err = secretMapHolder.Get(key[:len(key)-1])
		if err == nil {
			t.Fatal(err)
		}
	}
}

// illegal base64
func TestSecretMap7(t *testing.T) {
	testString1 := strings.Repeat("/", 10)
	testString2 := strings.Repeat("//", 10)
	testString3 := strings.Repeat("///", 10)
	testInputData := []string{testString1, testString2, testString3}
	secretMapHolder := NewSecretMapHolder()
	for _, data := range testInputData {
		_, err := secretMapHolder.Set([]byte(data))
		if err != nil {
			t.Fatal(err)
		}
		_, err = secretMapHolder.Get(testString1)
		if err == nil {
			t.Fatal(err)
		}
	}
}
