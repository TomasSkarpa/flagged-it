export const metaTranslations: Record<string, string> = {
	en: 'Learn about countries playing different game modes - guessing games, flag recognition, and shape identification.',
	es: 'Aprende sobre países jugando diferentes modos de juego: juegos de adivinanza, reconocimiento de banderas e identificación de formas.',
	fr: 'Apprenez-en sur les pays en jouant à différents modes de jeu - jeux de devinettes, reconnaissance de drapeaux et identification de formes.',
	de: 'Lernen Sie Länder kennen, indem Sie verschiedene Spielmodi spielen - Ratespiele, Flaggenerkennung und Formidentifikation.',
	nl: 'Leer over landen door verschillende spelmodi te spelen - raadspelletjes, vlagherkenning en vormidentificatie.',
	nb: 'Lær om land ved å spille forskjellige spillmoduser - gjettespill, flaggkjennskap og formidentifikasjon.',
	da: 'Lær om lande ved at spille forskellige spiltilstande - gættespil, flaggenkendelse og formidentifikation.',
	sv: 'Lär dig om länder genom att spela olika spellägen - gissningsspel, flaggkännedom och formidentifiering.',
	fi: 'Opi maista pelaamalla erilaisia pelimuotoja - arvauspelejä, lipuntunnistusta ja muotojen tunnistamista.',
	pt: 'Aprenda sobre países jogando diferentes modos de jogo - jogos de adivinhação, reconhecimento de bandeiras e identificação de formas.',
	tr: 'Farklı oyun modlarını oynayarak ülkeler hakkında bilgi edinin - tahmin oyunları, bayrak tanıma ve şekil tanımlama.',
	ro: 'Află despre țări jucând diferite moduri de joc - jocuri de ghicit, recunoaștere steaguri și identificare forme.',
	hu: 'Ismerjen meg országokat különböző játékmódok játszásával - tippelős játékok, zászlófelismerés és alakzatazonosítás.',
	hr: 'Saznajte više o zemljama igrajući različite načine igre - igre pogađanja, prepoznavanje zastava i identifikacija oblika.',
	cs: 'Poznejte země hraním různých herních režimů - hádací hry, rozpoznávání vlajek a identifikace tvarů.',
	sk: 'Zistite viac o krajinách hraním rôznych herných režimov - hádacie hry, rozpoznávanie vlajok a identifikácia tvarov.',
	pl: 'Poznaj kraje, grając w różne tryby gry - gry zgadywanki, rozpoznawanie flag i identyfikacja kształtów.',
	it: 'Scopri i paesi giocando diverse modalità di gioco - giochi di indovinelli, riconoscimento bandiere e identificazione forme.',
	id: 'Pelajari negara dengan memainkan berbagai mode permainan - permainan tebak-tebakan, pengenalan bendera dan identifikasi bentuk.',
	ms: 'Pelajari negara dengan memainkan pelbagai mod permainan - permainan tekaan, pengenalan bendera dan pengenalan bentuk.',
	fil: 'Matuto tungkol sa mga bansa sa pamamagitan ng paglalaro ng iba\'t ibang mga mode ng laro - mga laro ng paghula, pagkilala ng watawat at pagkilala ng hugis.',
	sw: 'Jifunze kuhusu nchi kwa kucheza njia tofauti za mchezo - michezo ya kukisia, utambuzi wa bendera na utambuzi wa umbo.',
	vi: 'Tìm hiểu về các quốc gia bằng cách chơi các chế độ trò chơi khác nhau - trò chơi đoán, nhận dạng cờ và nhận dạng hình dạng.'
};

export function getMetaDescription(locale: string = 'en'): string {
	return metaTranslations[locale] || metaTranslations.en;
}


