package siga

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"testing"
	"time"
)

func TestClient_MobileIDSigning_Succeeds(t *testing.T) {
	// given
	c := testClient(t)
	defer c.Close()

	ctx := context.Background()
	const session = "TestClient_MobileIDSigning_Succeeds"
	datafile, err := ReadDataFile("client_test.go")
	if err != nil {
		t.Fatal(err)
	}

	const person = "60001019906"
	const phone = "+37200000766"
	const message = "Automated testing"

	output, err := os.Create("testdata/mobile-id.asice")
	if err != nil {
		t.Fatal(err)
	}
	defer output.Close()

	// when
	if err = c.CreateContainer(ctx, session, datafile); err != nil {
		t.Fatal("create container:", err)
	}
	defer c.CloseContainer(ctx, session) // Attempt to clean-up SiGa regardless.

	if _, err = c.StartMobileIDSigning(ctx, session, person, phone, message); err != nil {
		t.Fatal("start signing:", err)
	}

	ticker := time.NewTicker(5 * time.Second)
	defer ticker.Stop()
	for range ticker.C {
		done, err := c.RequestMobileIDSigningStatus(ctx, session)
		if err != nil {
			t.Fatal("poll signing:", err)
		}
		if done {
			break
		}
	}

	if err = c.WriteContainer(ctx, session, output); err != nil {
		t.Fatal("write container:", err)
	}

	// then
	// The file written to testdata/mobile-id.asice should be externally
	// validated using e.g. DigiDoc4 Client.
}

func TestClient_UploadContainer_Succeeds(t *testing.T) {
	// given
	c := testClient(t)
	defer c.Close()

	ctx := context.Background()
	const session = "TestClient_UploadContainer_Succeeds"
	container, err := os.Open("testdata/mobile-id.asice")
	if err != nil {
		t.Fatal(err) // Will fail if TestClient_MobileIDSigning_Succeeds was skipped.
	}
	defer container.Close()

	// when
	err = c.UploadContainer(ctx, session, container)
	defer c.CloseContainer(ctx, session) // Attempt to clean-up SiGa regardless.

	// then
	if err != nil {
		t.Fatal("upload container:", err)
	}
}

// testClient return a client connected to the test server specified in
// testdata/siga.json. If no such file exists, then testClient skips the test.
//
// If the configuration contains any Ignite servers, then those are used for
// storage, otherwise memory storage is used instead.
func testClient(t *testing.T) Client {
	bytes, err := ioutil.ReadFile("testdata/siga.json")
	switch {
	case err == nil:
	case os.IsNotExist(err):
		t.Skip("no test server configured in testdata/siga.json")
	default:
		t.Fatal(err)
	}
	var conf Conf
	if err := json.Unmarshal(bytes, &conf); err != nil {
		t.Fatal(err)
	}

	// Väljasta kontrolliks sisseloetud konf.
	fmt.Println(prettyPrint(conf))
	
	/* if len(conf.Ignite.Servers) > 0 {
		c, err := NewClient(conf)
		if err != nil {
			t.Fatal(err)
		}
		return c
	} */

	c, err := newClientWithoutStorage(conf)
	if err != nil {
		t.Fatal(err)
	}
	c.storage = newMemStorage()
	return c
}

// prettyPrint koostab struct-st i printimisvalmis sõne.
func prettyPrint(i interface{}) string {
	s, _ := json.MarshalIndent(i, "", "\t")
	return string(s)
}

