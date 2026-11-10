package i18n

import "fmt"

// catalog holds every interface string in both languages. Keys are grouped
// by screen. Data texts (events, evidence) live in the YAML files instead.
var catalog = map[string]Text{
	// Title screen.
	"title.tagline":   T("Пошаговый рогалик о том, как дожить до 85. Каждый риск — из научных данных.", "A turn-based roguelike about living to 85. Every risk comes from real evidence."),
	"title.new":       T("Новая жизнь", "New life"),
	"title.graveyard": T("Кладбище", "Graveyard"),
	"title.about":     T("Об игре и источниках", "About & sources"),
	"title.lang":      T("Язык: Русский", "Language: English"),
	"title.quit":      T("Выход", "Quit"),
	"title.best":      T("Лучший результат: %d очков (%d лет)", "Best score: %d points (age %d)"),
	"title.lastdead":  T("Последние смерти:", "Recent deaths:"),
	"title.based":     T("По книге «HowToLiveBetter» (772 рекомендации, уровни доказательности A/B/C) и таблицам смертности ВОЗ.", "Based on the book \"HowToLiveBetter\" (772 recommendations, evidence grades A/B/C) and WHO life tables."),

	// Setup screen.
	"setup.title":      T("НОВАЯ ЖИЗНЬ", "A NEW LIFE"),
	"setup.region":     T("Страна (таблица смертности)", "Country (life table)"),
	"setup.sex":        T("Пол", "Sex"),
	"setup.profession": T("Профессия", "Profession"),
	"setup.seed":       T("Зерно", "Seed"),
	"setup.start":      T("▶ Родиться (точнее, стать взрослым)", "▶ Be born (well, turn eighteen)"),
	"setup.baseline":   T("Шанс дожить с 18 до 85 по таблице ВОЗ: %s", "WHO life-table chance of living from 18 to 85: %s"),
	"setup.hint":       T("j/k — поле · h/l — значение · Enter — начать · Esc — назад", "j/k field · h/l change · Enter start · Esc back"),
	"sex.male":         T("Мужской", "Male"),
	"sex.female":       T("Женский", "Female"),
	"setup.seed.fixed": T("%d (задано)", "%d (fixed)"),

	// Status bar.
	"bar.age":    T("ВОЗРАСТ", "AGE"),
	"bar.hp":     T("HP", "HP"),
	"bar.sanity": T("SANITY", "SANITY"),
	"bar.wealth": T("КАПИТАЛ", "WEALTH"),
	"bar.will":   T("ВОЛЯ", "WILL"),

	// Event screen.
	"event.year":       T("Год %d", "Year %d"),
	"event.quiet":      T("Тихий год", "A quiet year"),
	"event.quiet.text": T("Ничего особенного не произошло. Жизнь шла своим чередом.", "Nothing special happened. Life went on."),
	"event.continue":   T("Прожить этот год", "Live through the year"),
	"event.conditions": T("Состояния:", "Conditions:"),
	"event.none":       T("нет", "none"),
	"event.nomoney":    T("не хватает денег", "not enough money"),
	"event.nowill":     T("не хватает воли", "not enough willpower"),
	"event.cost.will":  T("воля", "will"),
	"event.cost.time":  T("%s ч", "%s h"),
	"event.free":       T("бесплатно", "free"),
	"event.keys":       T("↑↓/jk выбор · Enter/l решить · 1-4 быстро · L язык · T тема · ? помощь · q выход", "↑↓/jk select · Enter/l choose · 1-4 quick · L lang · T theme · ? help · q quit"),

	// Category names.
	"cat.health":    T("здоровье", "health"),
	"cat.chronic":   T("образ жизни", "lifestyle"),
	"cat.emergency": T("экстренно", "emergency"),
	"cat.pharmacy":  T("аптечка", "medicines"),
	"cat.safety":    T("безопасность", "safety"),
	"cat.money":     T("деньги", "money"),
	"cat.legal":     T("закон", "law"),
	"cat.work":      T("работа", "work"),
	"cat.family":    T("близкие", "family"),
	"cat.mind":      T("психика", "mind"),
	"cat.travel":    T("поездки", "travel"),
	"cat.life":      T("жизнь", "life"),

	// Outcome screen.
	"out.chose":                  T("Вы выбрали:", "You chose:"),
	"out.changes":                T("Изменения:", "Changes:"),
	"out.risks":                  T("Скрытые риски:", "Hidden risks:"),
	"out.years":                  T("на %d г.", "for %d y"),
	"out.added":                  T("Новое состояние:", "New condition:"),
	"out.removed":                T("Больше нет:", "No longer:"),
	"out.survived":               T("Пронесло: шанс погибнуть был %s.", "You survived a %s chance of dying."),
	"out.disaster":               T("Случилось худшее (шанс был %s):", "The worst happened (chance was %s):"),
	"out.dodged":                 T("Беда обошла стороной: шанс был %s.", "Disaster passed you by: the chance was %s."),
	"out.evidence":               T("ЧТО ГОВОРИТ НАУКА", "WHAT THE EVIDENCE SAYS"),
	"out.chapter":                T("Глава %d «%s» · пункт %d", "Chapter %d \"%s\" · item %d"),
	"out.grade":                  T("Уровень доказательности %s", "Evidence grade %s"),
	"out.contested":              T("есть споры", "contested"),
	"out.source":                 T("Источник:", "Source:"),
	"out.book":                   T("В книге:", "In the book:"),
	"out.yearend":                T("ИТОГИ ГОДА", "THE YEAR IN REVIEW"),
	"out.income":                 T("Доход после расходов:", "Income after living costs:"),
	"out.noincome":               T("Дохода нет", "No income"),
	"out.interest":               T("Проценты по долгам:", "Interest on debt:"),
	"out.upkeep":                 T("Возраст и привычки:", "Age and habits:"),
	"out.risk":                   T("Риск умереть в этом году: %s (×%s к сверстникам)", "Risk of dying this year: %s (×%s vs your peers)"),
	"out.passed":                 T("Ежегодная проверка пройдена — вам %d.", "Annual check passed — you are now %d."),
	"out.arrest":                 T("Прошлое догнало вас: арест (шанс был %s).", "Your past caught up with you: arrested (chance was %s)."),
	"out.expired":                T("Закончилось:", "Ended:"),
	"out.continue":               T("Enter — следующий год · j/k — прокрутка", "Enter next year · j/k scroll"),
	"out.dead":                   T("Enter — к протоколу вскрытия", "Enter — to the autopsy report"),
	"out.victory":                T("Enter — дальше", "Enter — continue"),
	"risk.cardiovascular_risk":   T("сердце и сосуды", "heart & vessels"),
	"risk.oncology_risk":         T("онкология", "cancer"),
	"risk.toxicology_liver_risk": T("печень и токсины", "liver & toxins"),
	"risk.trauma_accident_risk":  T("травмы", "injuries"),
	"risk.legal_prison_risk":     T("закон", "the law"),
	"stat.health":                T("Здоровье", "Health"),
	"stat.sanity":                T("Психика", "Sanity"),
	"stat.wealth":                T("Капитал", "Wealth"),
	"stat.will":                  T("Воля", "Will"),

	// Morgue.
	"morgue.title":     T("ПРОТОКОЛ ВСКРЫТИЯ № %d", "AUTOPSY REPORT No. %d"),
	"morgue.patient":   T("Пациент:", "Patient:"),
	"morgue.age":       T("Возраст на момент смерти:", "Age at death:"),
	"morgue.age.value": T("%d лет", "%d years"),
	"morgue.cause":     T("Причина смерти:", "Cause of death:"),
	"morgue.group":     T("Группа причин (ВОЗ):", "Cause group (WHO):"),
	"morgue.instant":   T("на месте, сразу после решения", "on the spot, right after a decision"),
	"morgue.collapse":  T("здоровье исчерпано до нуля", "health ran down to zero"),
	"morgue.annual":    T("ежегодная проверка смертности (шанс в тот год: %s)", "annual mortality check (chance that year: %s)"),
	"morgue.how":       T("Обстоятельства:", "Circumstances:"),
	"morgue.decision":  T("РЕШЕНИЕ, КОТОРОЕ К ЭТОМУ ПРИВЕЛО", "THE DECISION THAT LED HERE"),
	"morgue.at":        T("в возрасте %d", "at age %d"),
	"morgue.factor":    T("Повысило риск этой причины в %s раза", "Raised the risk of this cause %s-fold"),
	"morgue.natural":   T("Ни одно решение не повысило риск этой причины: таков базовый риск по таблице ВОЗ для вашего возраста.", "No decision raised the risk of this cause: this is the baseline WHO life-table risk for your age."),
	"morgue.science":   T("НАУЧНОЕ ОБОСНОВАНИЕ", "SCIENTIFIC BASIS"),
	"morgue.lost":      T("Потеряно лет жизни (по таблице ВОЗ):", "Years of life lost (WHO table):"),
	"morgue.short":     T("Не дожили до 85:", "Short of 85 by:"),
	"morgue.years":     T("%d лет", "%d years"),
	"morgue.peers":     T("Вы пережили %s ровесников из той же таблицы.", "You outlived %s of your peers in the same table."),
	"morgue.mods":      T("ВСКРЫТЫЕ МОДИФИКАТОРЫ РИСКА", "HIDDEN RISK MODIFIERS, REVEALED"),
	"morgue.stats":     T("СТАТИСТИКА ЖИЗНИ", "LIFE STATISTICS"),
	"morgue.decisions": T("Решения: мудрых %d · нейтральных %d · рискованных %d", "Decisions: wise %d · neutral %d · risky %d"),
	"morgue.hours":     T("Часов вложено в себя: %s · Потрачено: %s", "Hours invested in yourself: %s · Money spent: %s"),
	"morgue.prison":    T("Лет в заключении: %d · Арестов: %d", "Years in prison: %d · Arrests: %d"),
	"morgue.peak":      T("Пиковый капитал: %s", "Peak wealth: %s"),
	"morgue.sources":   T("Источников из книги на вашем пути: %d", "Book items along your way: %d"),
	"morgue.score":     T("ИТОГОВЫЙ СЧЁТ: %d", "FINAL SCORE: %d"),
	"morgue.seed":      T("Зерно для повтора: %d (survive85 -seed %d)", "Seed to replay: %d (survive85 -seed %d)"),
	"morgue.keys":      T("Enter — новая жизнь · r — та же судьба заново · g — кладбище · q — выход · j/k — прокрутка", "Enter new life · r replay this fate · g graveyard · q quit · j/k scroll"),
	"morgue.rip":       T("ПОКОЙСЯ С МИРОМ", "REST IN PEACE"),

	// Victory.
	"victory.title":   T("ВЫ ДОЖИЛИ ДО 85", "YOU MADE IT TO 85"),
	"victory.text":    T("Вы пережили %s ровесников. Таблица ВОЗ давала вам %s шансов.", "You outlived %s of your peers. The WHO table gave you a %s chance."),
	"victory.top":     T("САМЫЕ ВЛИЯТЕЛЬНЫЕ РЕШЕНИЯ", "THE DECISIONS THAT MATTERED MOST"),
	"victory.keys":    T("Enter — новая жизнь · r — та же судьба заново · g — кладбище · q — выход", "Enter new life · r replay this fate · g graveyard · q quit"),
	"victory.epitaph": T("Ещё не надгробие, а торт с 85 свечами.", "Not a tombstone yet — a cake with 85 candles."),

	// Graveyard.
	"grave.title":   T("КЛАДБИЩЕ ПРОЖИТЫХ ЖИЗНЕЙ", "GRAVEYARD OF LIVES LIVED"),
	"grave.empty":   T("Пока пусто. Это ненадолго.", "Empty so far. Not for long."),
	"grave.victory": T("дожил(а) до 85", "lived to 85"),
	"grave.keys":    T("Esc — назад · j/k — прокрутка", "Esc back · j/k scroll"),

	// About.
	"about.title": T("ОБ ИГРЕ И ИСТОЧНИКАХ", "ABOUT & SOURCES"),
	"about.body": T(`survive85 — открытая игра (лицензия MIT). Вы живёте с 18 до 85 лет, по одному событию в год.

КАК СЧИТАЕТСЯ СМЕРТНОСТЬ
Каждый год движок берёт вероятность умереть для вашего возраста, пола и страны из таблиц смертности ВОЗ (2019) и делит её по группам причин по данным Global Health Estimates ВОЗ: сердце и сосуды, онкология, печень и отравления, травмы, суицид, прочее. Каждая группа умножается на ваш скрытый модификатор — так же, как работает отношение рисков (HR) в исследованиях. Здоровье ниже типичного для возраста повышает риск болезней, психика ниже 50 — риск суицида.

ОТКУДА ЦИФРЫ
Все рекомендации, уровни доказательности (A/B/C), отношения рисков (HR/OR/RR), DOI и цитаты взяты из книги «HowToLiveBetter» (eternity4719, CC BY 4.0). Русские и английские цитаты — из перевода dlgrv/HowToLiveBetter. Абсолютные шансы мгновенных проверок помечены в данных как игровые приближения: из книги берётся соотношение между вариантами.

ЛИЦЕНЗИИ
Код: MIT. Данные: CC BY 4.0 — © eternity4719, https://github.com/eternity4719/HowToLiveBetter (изменения: пункты отобраны, процитированы и превращены в игровые события). Перевод: https://github.com/dlgrv/HowToLiveBetter. Таблицы смертности и причины смерти: Всемирная организация здравоохранения, Global Health Observatory и Global Health Estimates.

ВАЖНО
Это игра, а не медицинская или юридическая консультация. Если вам плохо — звоните 112 или на телефон доверия и поговорите с близкими.`,
		`survive85 is an open-source game (MIT licence). You live from 18 to 85, one event per year.

HOW MORTALITY WORKS
Every year the engine takes the probability of dying at your age, sex and country from WHO life tables (2019) and splits it into cause groups using WHO Global Health Estimates: heart & vessels, cancer, liver & poisoning, injuries, suicide, other. Each group is multiplied by your hidden modifier — exactly how a hazard ratio (HR) works in studies. Health below what is typical for your age raises disease risk; sanity below 50 raises suicide risk.

WHERE THE NUMBERS COME FROM
Every recommendation, evidence grade (A/B/C), risk ratio (HR/OR/RR), DOI and quote comes from the book "HowToLiveBetter" (eternity4719, CC BY 4.0). English and Russian quotes come from the dlgrv/HowToLiveBetter translation. Absolute chances of instant checks are marked in the data as game approximations: the book supplies the ratio between choices.

LICENCES
Code: MIT. Data: CC BY 4.0 — © eternity4719, https://github.com/eternity4719/HowToLiveBetter (changes: items selected, quoted and turned into game events). Translation: https://github.com/dlgrv/HowToLiveBetter. Life tables and causes of death: World Health Organization, Global Health Observatory and Global Health Estimates.

IMPORTANT
This is a game, not medical or legal advice. If you are struggling, call your local emergency number (112 / 911) or a crisis line, and talk to someone close to you.`),
	"about.keys": T("Esc — назад · j/k — прокрутка", "Esc back · j/k scroll"),

	// Help overlay.
	"help.title": T("УПРАВЛЕНИЕ", "CONTROLS"),
	"help.body": T(`↑ ↓  или  k j     выбор
Enter Space l     подтвердить
h  Esc            назад
1 2 3 4           быстрый выбор варианта
PgUp PgDn         прокрутка длинных экранов
L                 русский / английский
T                 сменить тему (Nord, Cyberpunk, Monokai)
?                 эта подсказка
q  Ctrl+C         выход

Сила воли (◆) тратится на дисциплинированные решения и восстанавливается каждый год.
Скрытые модификаторы риска вы увидите только в протоколе вскрытия.`,
		`↑ ↓  or  k j      move
Enter Space l     confirm
h  Esc            back
1 2 3 4           quick-pick a choice
PgUp PgDn         scroll long screens
L                 Russian / English
T                 cycle theme (Nord, Cyberpunk, Monokai)
?                 this help
q  Ctrl+C         quit

Willpower (◆) is spent on disciplined choices and recovers every year.
Hidden risk modifiers are revealed only in the autopsy report.`),
	"help.close": T("любая клавиша — закрыть", "any key to close"),

	"small": T("Окно слишком маленькое: нужно хотя бы %d×%d (сейчас %d×%d).", "The terminal is too small: need at least %d×%d (now %d×%d)."),
	"theme": T("Тема: %s", "Theme: %s"),
}

// S returns the interface string for key in language l. Unknown keys are
// returned as-is so a missing translation is visible rather than blank.
func S(l Lang, key string) string {
	if t, ok := catalog[key]; ok {
		return t.Get(l)
	}
	return key
}

// F formats the interface string for key with fmt.Sprintf.
func F(l Lang, key string, args ...any) string {
	return fmt.Sprintf(S(l, key), args...)
}

// Keys lists every catalog key (for tests).
func Keys() []string {
	out := make([]string, 0, len(catalog))
	for k := range catalog {
		out = append(out, k)
	}
	return out
}

// Lookup returns the Text for key.
func Lookup(key string) (Text, bool) {
	t, ok := catalog[key]
	return t, ok
}
