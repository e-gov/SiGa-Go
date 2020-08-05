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
    fetch('https://localhost:8080/p1', {
      method: 'POST',
      headers: {
        'content-type': 'application/json'
      },
      body: JSON.stringify({
        tekst: document.getElementById("Tekstisisestusala").innerText
      })
    })
      .then(response => {
        console.log(response)
      })
      .catch(err => {
        console.log(err)
      })
  });
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