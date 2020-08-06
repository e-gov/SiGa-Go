'use strict';

// Peaprogramm. Mitmesugused algväärtustamised.
function init() {

  seaTeabepaanideKasitlejad();
  seaTekstinupukasitlejad();

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

// Sea sisestatava teksti käsitlejad ('Uus', 'Vaheta pooled', 'Salvesta').
function seaTekstinupukasitlejad() {

  $('#Allkirjastanupp').click(() => {

    // HwCrypto töökorras oleku kontroll.
    if (!window.hwcrypto.use('auto')) {
      console.error("loeSert: hwcrypto BE valik ebaõnnestus.");
    }
  
    window.hwcrypto.debug()
      .then(
        (response) => {
          console.log("loeSert: Debug: " + response);
          // Asünk läheb siin katki? Kuid kasutaja on aeglane..
        },
        (err) => {
          console.log("loeSert: debug() failed: " + err);
          return;
        });
  
    var options = { lang: 'et' };
    // filter: 'AUTH' valik salgamatuseta serdid
  
    // Päri sert ja saada see koos allkirjastatava tekstiga serveripoolele.
    window.hwcrypto.getCertificate(options)
      .then(
        function (response) {
          var certPEM = hexToPem(response.hex);
          // var certDER = response.encoded;
          console.log("loeSert: Sert loetud:\n");
          ///////////////////////
          document.getElementById("Teatetekst").innerText = certPEM;

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
            .then(response => {
              // TODO: Lisada vastuse töötlemine
              console.log("loeSert: Vastus: ", response)
            })
            .catch(err => {
              console.log("loeSert: Viga P1 saatmisel: ", err)
            })
            ///////////////////////
        },
        function (err) {
          console.error("loeSert: Serdi lugemine ebaõnnestus. ",
            "Kontrolli, kas ID-kaart on lugejas. : "
            + err);
          document.getElementById("Teatetekst").innerText =
          "Serdita ei saa allkirjastada"
          return;
        }
      );
  });
}

// loeSert küsib kasutajalt allkirjastamisserdi valimist. Tagastab serdi
// või null. NB! Asünkroonne.
function loeSert() {
}

// kuvaTeade kuvab teate.
function kuvaTeade(t) {
  $('#Teatetekst').text(t);
  $('#Teatepaan').removeClass('peidetud');
}

// Märkmed

// Using Fetch API
// https://developer.mozilla.org/en-US/docs/Web/API/Fetch_API/Using_Fetch

// Fetch (vigane näide!)
// https://flaviocopes.com/how-to-post-api-javascript/