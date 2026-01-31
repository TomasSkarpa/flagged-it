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
	vi: 'Tìm hiểu về các quốc gia bằng cách chơi các chế độ trò chơi khác nhau - trò chơi đoán, nhận dạng cờ và nhận dạng hình dạng.',
	ar: 'تعرف على البلدان من خلال لعب أنماط ألعاب مختلفة - ألعاب التخمين، التعرف على الأعلام وتحديد الأشكال.',
	el: 'Μάθετε για τις χώρες παίζοντας διαφορετικούς τρόπους παιχνιδιού - παιχνίδια εικασίας, αναγνώριση σημαιών και αναγνώριση σχημάτων.',
	he: 'למד על מדינות על ידי משחק במצבי משחק שונים - משחקי ניחוש, זיהוי דגלים וזיהוי צורות.',
	hi: 'विभिन्न गेम मोड खेलकर देशों के बारे में जानें - अनुमान लगाने वाले खेल, झंडा पहचान और आकार पहचान।',
	ja: 'さまざまなゲームモードをプレイして国について学ぶ - 推測ゲーム、旗の認識、形状の識別。',
	ko: '다양한 게임 모드를 플레이하여 국가에 대해 알아보세요 - 추측 게임, 깃발 인식 및 모양 식별.',
	ru: 'Узнавайте о странах, играя в различные игровые режимы - игры на угадывание, распознавание флагов и определение форм.',
	th: 'เรียนรู้เกี่ยวกับประเทศต่างๆ โดยเล่นโหมดเกมที่แตกต่างกัน - เกมทาย, การจดจำธงและการระบุรูปร่าง.',
	uk: 'Дізнавайтеся про країни, граючи в різні ігрові режими - ігри на вгадування, розпізнавання прапорів та визначення форм.',
	zh: '通过玩不同的游戏模式了解国家 - 猜谜游戏、旗帜识别和形状识别。'
};

export function getMetaDescription(locale: string = 'en'): string {
	return metaTranslations[locale] || metaTranslations.en;
}

const SITE_URL = 'https://flaggedit.app';
const DEFAULT_OG_IMAGE = `${SITE_URL}/assets/favicon/apple-touch-icon.png`;

export function getOGTitle(title?: string, locale: string = 'en'): string {
	if (title) {
		return `${title} - Flagged It`;
	}
	return 'Flagged It - Learn Countries Through Fun Guessing Games';
}

export function getOGDescription(locale: string = 'en'): string {
	return getMetaDescription(locale);
}

export function getOGImage(image?: string): string {
	return image || DEFAULT_OG_IMAGE;
}

export function getOGUrl(path?: string): string {
	if (!path) {
		return SITE_URL;
	}
	// Ensure path starts with /
	const cleanPath = path.startsWith('/') ? path : `/${path}`;
	return `${SITE_URL}${cleanPath}`;
}

