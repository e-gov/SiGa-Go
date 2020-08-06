'use strict';

// Peaprogramm. Mitmesugused algväärtustamised.
function init() {

  seaTeabepaanideKasitlejad();
  seaNupukasitlejad();

}

// Infopaani käsitlejad: avamine, sulgemine.
function seaTeabepaanideKasitlejad() {
  $('#Info').click(() => {
    $('#Infopaan').removeClass('peidetud');
    $('#Info').addClass('disabled');
  });

  $('#InfopaanSulge').click(() => {
    $('#Infopaan').addClass('peidetud');
    $('#Info').removeClass('disabled');
  });

  $('#TeatepaanSulge').click(() => {
    $('#Teatepaan').addClass('peidetud');
  });
}

// XXX: Õige viis serdi edasiandmiseks
var certToUse;

// seaNupukasitlejad määrab "Allkirjasta ID-kaardiga" ja "Allkirjasta m-ID-ga"
// käitumise.
function seaNupukasitlejad() {

  $('#IDkaartNupp').click(() => {

    // HwCrypto töökorras oleku kontroll.
    if (!window.hwcrypto.use('auto')) {
      console.error("ID-kaardiga: hwcrypto BE valik ebaõnnestus.");
    }
   
    var options = { lang: 'et' };
    // filter: 'AUTH' valik salgamatuseta serdid
  
    // Päri sert ja saada see koos allkirjastatava tekstiga serveripoolele.
    window.hwcrypto.getCertificate(options)
      .then(
        (certificate) => {
          var certPEM = hexToPem(certificate.hex);
          // var certDER = certificate.encoded;
          console.log("ID-kaardiga: Sert loetud:\n");
          certToUse = certificate;
          ///////////////////////
          kuvaTeade(certPEM, false);

          // Saada allkirjastatav tekst ja sert serveripoolele.
          fetch('https://localhost:8080/p1', {
            method: 'POST',
            headers: {
              'content-type': 'application/json'
            },
            body: JSON.stringify({
              tekst: document.getElementById("Tekstisisestusala").innerText,
              sert: certPEM 
            })
          })
            // Loe vastus sisse, JSON-na
            .then(response => { 
              console.log("Vastus saadud!");
              response.json() 
              .then(data => { 
                console.log("Vastuse keha: ", data);
                IDkaardiga2(data.hash, data.algo)
              })
            })
            .catch(err => {
              console.log("ID-kaardiga: Viga P1 saatmisel: ", err)
            })
        },
        function (err) {
          console.error("ID-kaardiga: Serdi lugemine ebaõnnestus. ",
            "Kontrolli, kas ID-kaart on lugejas. : "
            + err);
          kuvaTeade("Serdi lugemine ebaõnnestus. Kontrolli, kas ID-kaart on lugejas.",
            true);
          return;
        }
      );
  });

  $('#mIDNupp').click(() => {
    // Saada allkirjastatav tekst, isikukood ja mobiilinr serveripoolele.
    fetch('https://localhost:8080/mid', {
      method: 'POST',
      headers: {
        'content-type': 'application/json'
      },
      body: JSON.stringify({
        isikukood: "60001019906",
        nr: "37200000766",
        tekst: document.getElementById("Tekstisisestusala").innerText
      })
    })
      // Loe vastus sisse, JSON-na
      .then(response => { 
        console.log("Vastus saadud!");
        response.json() 
        .then(data => { 
          console.log("Vastuse keha: ", data);
          kuvaTeade("Allkirjastamine edukas", false)
        })
      })
      .catch(err => {
        console.log("m-ID-ga: Viga P1 saatmisel: ", err)
        kuvaTeade("Päring ebaõnnestus", true)
      })

  });
}

// IDkaardiga2 teeb allkirjastamise teise osa: PIN2 küsimine jne.
function IDkaardiga2(hash, algo) {
  console.log("Alustan allkirjastamise 2. osa")
  var plainHash = new Uint8Array(base64js.toByteArray(hash));
  window.hwcrypto.sign(
    certToUse,
    { type: algo, value: plainHash },
    { lang: 'et'})
    .then(
      (signature) => {
        console.log("Allkirjastamine õnnestus")
        // Saada signature.value serveripoolele
        saadaAllkiri(signature.value)
      },
      (err) => {
        console.log("Allkirjastamine ebaõnnestus: ", err)
      }
    );
}

// saadaAllkiri saadab allkirja serveripoolele ja viib ID-kaardiga allkirjastamise
// sirvikupoolel lõpuni.
function saadaAllkiri(signatureValue) {
  // Teisenda Base64 kujule
  var s = base64js.fromByteArray(signatureValue);
  fetch('https://localhost:8080/p2', {
    method: 'POST',
    headers: {
      'content-type': 'application/json'
    },
    body: JSON.stringify({
      allkiri: s
    })
  })
    // Loe vastus sisse, JSON-na
    .then(response => { 
      console.log("P2 vastus saadud!");
      response.json() 
      .then(data => { 
        console.log("P2 Vastuse keha: ", data);
        // Töötle vastus
        if (data.error !== "") {
          kuvaTeade(data.error, true);
        } else {
          kuvaTeade("Allkiri edukalt antud", false);
        }
      })
    })
    .catch(err => {
      console.log("ID-kaardiga: Viga P2 saatmisel: ", err)
    })
}

// kuvaTeade kuvab teate.
function kuvaTeade(t, error) {
  $('#Teatetekst').text(t);
  if (error) {
    $('#Teatepaan').addClass('tomato').removeClass('green');
  } else {
    $('#Teatepaan').addClass('green').removeClass('tomato');
  }
  $('#Teatepaan').removeClass('peidetud');
}

// Märkmed

// Using Fetch API
// https://developer.mozilla.org/en-US/docs/Web/API/Fetch_API/Using_Fetch

// Fetch (vigane näide!)
// https://flaviocopes.com/how-to-post-api-javascript/