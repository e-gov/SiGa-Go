package siga

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"time"

	"github.com/e-gov/SiGa-Go/abi"
)

// Example_MobileIDSigning demonstreerib m-ID-ga
// allkirjastamist.
// Selleks:
// 1) moodustab Riigi allkirjastamisteenuse (SiGa) poole pöördumise
// HTTPS kliendi (TestClient)
// 2) alustab SiGa-ga seanssi (session)
// 3) valib allkirjastatava faili (testdata/example_datafile.txt)
// 4) koostab konteineri (CreateContainer)
// 5) teeb m-ID-ga allkirjastamise alustamise päringu
// (StartMobileIDSigning)
// 6) teeb m-ID-ga allkirjastamise seisundipäringud
// (RequestMobileIDSigningStatus)
// 7) salvestab konteineri (WriteContainer), faili
// testdata/mobile-id.asice
// 8) suleb HTTPS kliendi (Close).
func Example_MobileIDSigning() {
	// given
	c := TestClient()
	defer c.Close()

	ctx := context.Background()
	const session = "TestClient_MobileIDSigning_Succeeds"
	datafile, err := ReadDataFile("testdata/example_datafile.txt")
	if err != nil {
		fmt.Println("Example_MobileIDSigning: ", err)
		os.Exit(1)
	}

	const person = "60001019906"
	const phone = "+37200000766"
	const message = "Automated testing"

	output, err := os.Create("testdata/mobile-id.asice")
	if err != nil {
		fmt.Println("Example_MobileIDSigning: ", err)
		os.Exit(1)
	}
	defer output.Close()

	// when
	if err = c.CreateContainer(ctx, session, datafile); err != nil {
		fmt.Println("Example_MobileIDSigning: ", err)
		os.Exit(1)
	}
	defer c.CloseContainer(ctx, session) // Attempt to clean-up SiGa regardless.

	if _, err = c.StartMobileIDSigning(ctx, session, person, phone, message); err != nil {
		fmt.Println("Example_MobileIDSigning: ", err)
		os.Exit(1)
	}

	ticker := time.NewTicker(5 * time.Second)
	defer ticker.Stop()
	for range ticker.C {
		done, err := c.RequestMobileIDSigningStatus(ctx, session)
		if err != nil {
			fmt.Println("Example_MobileIDSigning: ", err)
			os.Exit(1)
		}
		if done {
			break
		}
	}

	if err = c.WriteContainer(ctx, session, output); err != nil {
		fmt.Println("Example_MobileIDSigning: ", err)
		os.Exit(1)
	}

	fmt.Println("Example_MobileIDSigning: Allkiri moodustatud")

	// then
	// The file written to testdata/mobile-id.asice should be externally
	// validated using e.g. DigiDoc4 Client.
}

/* func TestClient_UploadContainer_Succeeds(t *testing.T) {
	// given
	c := TestClient(t)
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
*/

// TestClient return a client connected to the test server specified in
// testdata/siga.json. If no such file exists, then testClient skips the test.
//
// If the configuration contains any Ignite servers, then those are used for
// storage, otherwise memory storage is used instead.
func TestClient() Client {
	bytes, err := ioutil.ReadFile("testdata/siga.json")
	if err != nil {
		fmt.Println("TestClient: Viga seadistusfaili lugemisel: ", err)
		os.Exit(1)
	}

	var conf Conf
	if err := json.Unmarshal(bytes, &conf); err != nil {
		fmt.Println("TestClient: Viga seadistusfaili parsimisel: ", err)
		os.Exit(1)
	}

	// Väljasta kontrolliks sisseloetud konf.
	fmt.Println(abi.PrettyPrint(conf))
	
	/* if len(conf.Ignite.Servers) > 0 {
		c, err := NewClient(conf)
		if err != nil {
			t.Fatal(err)
		}
		return c
	} */

	c, err := newClientWithoutStorage(conf)
	if err != nil {
		fmt.Println("TestClient: ", err)
		os.Exit(1)
	}
	c.storage = newMemStorage()
	return c
}

// Märkmed

// prettyPrint
// https://stackoverflow.com/questions/24512112/how-to-print-struct-variables-in-console
