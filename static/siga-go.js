'use strict';

// Peaprogramm. Mitmesugused algväärtustamised.
function init() {

  seaTeabepaanideKasitlejad();
  seaRedaktoriKasitlejad();
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

  $('#Uusnupp').click(() => {
    $('#Tekst').focus();
  });

  $('#Allkirjastanupp').click(() => {
  });
}

// kuvaTeade kuvab teate.
function kuvaTeade(t) {
  $('#Teatetekst').text(t);
  $('#Teatepaan').removeClass('peidetud');
}
