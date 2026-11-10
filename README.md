<div align="center">

```
███████╗██╗   ██╗██████╗ ██╗   ██╗██╗██╗   ██╗███████╗    █████╗ ███████╗
██╔════╝██║   ██║██╔══██╗██║   ██║██║██║   ██║██╔════╝██╗██╔══██╗██╔════╝
███████╗██║   ██║██████╔╝██║   ██║██║██║   ██║█████╗  ╚═╝╚█████╔╝███████╗
╚════██║██║   ██║██╔══██╗╚██╗ ██╔╝██║╚██╗ ██╔╝██╔══╝  ██╗██╔══██╗╚════██║
███████║╚██████╔╝██║  ██║ ╚████╔╝ ██║ ╚████╔╝ ███████╗╚═╝╚█████╔╝███████║
╚══════╝ ╚═════╝ ╚═╝  ╚═╝  ╚═══╝  ╚═╝  ╚═══╝  ╚══════╝    ╚════╝ ╚══════╝
```

**A turn-based terminal roguelike about living from 18 to 85.**
**Every risk, every death and every number comes from real evidence.**

[![CI](https://github.com/funstorm15/survive85/actions/workflows/ci.yml/badge.svg)](https://github.com/funstorm15/survive85/actions/workflows/ci.yml)
[![Release](https://img.shields.io/github/v/release/funstorm15/survive85?sort=semver)](https://github.com/funstorm15/survive85/releases)
[![Go](https://img.shields.io/badge/Go-1.24%2B-00ADD8?logo=go)](go.mod)
[![Code: MIT](https://img.shields.io/badge/code-MIT-blue)](LICENSE)
[![Data: CC BY 4.0](https://img.shields.io/badge/data-CC%20BY%204.0-lightgrey)](data/LICENSE.md)
[![AUR](https://img.shields.io/aur/version/survive85?logo=archlinux)](https://aur.archlinux.org/packages/survive85)

**English** · [Русский](README.ru.md)

</div>

---

You turn 18. Every turn is one year of your life. Every year brings one
situation — a ride with a friend who "only had two beers", a cold-medicine
sachet next to a strip of paracetamol, a friend who needs a loan guarantor, a
boss who screams at you in front of everyone — and two to four ways to act.

Behind the scenes, an actuarial engine multiplies the **real WHO probability
of dying at your age, sex and country** by the **hazard ratios of your
decisions**, taken from the evidence-graded book
[**HowToLiveBetter**](https://github.com/eternity4719/HowToLiveBetter)
(772 recommendations, grades A/B/C, meta-analyses, cohort studies and RCTs).
When you die, you get an **autopsy report**: the cause, the decision that led
there, the study behind it with its DOI, and the years of life you lost.

Reach 85 and you win. Russian men reach it in 15.7% of cases. Good luck.

```
 [AGE: 34/85] | HP: [████████░░] 80% | SANITY: [██████░░░░] 60% | WEALTH: $6,160 | WILL: ◆◆◆◇◇◇◇◇◇◇
 ──────────────────────────────────────────────────────────────────────────────────────────────────
 ℞ A cold and a 39°C fever                                                   medicines · Year 34

 You're down with a bad flu. Your head is splitting and the thermometer says 39°C. The medicine
 cabinet holds a sachet of a hot lemon cold remedy and a strip of 500 mg paracetamol tablets.
 There's an important meeting tomorrow.

  ▸ 1. Drink the sachet and take two paracetamol on top — to be sure
    2. Read the label, spot the paracetamol in it and take only one
       will ◆
    3. Just the tablets — plus a beer to sweat it out
```

```
 AUTOPSY REPORT No. 11
 ╔═══════════════════════════════════════════════════════════════════════════════════════╗
 ║  Patient:                    Software developer, male, Russia                         ║
 ║  Age at death:               57 years                                                 ║
 ║  Cause of death:             Chronic heart failure                                    ║
 ║  Circumstances:              annual mortality check (chance that year: 8.31%)         ║
 ║                                                                                       ║
 ║  THE DECISION THAT LED HERE                                                           ║
 ║  Survived a heart attack at age 54 — raised the risk of this cause 2.00-fold          ║
 ║  SCIENTIFIC BASIS   Chapter 13, item 7 · [B] · DOI 10.1093/ehjqcco/qcaa037            ║
 ║  Years of life lost (WHO table): 18.7 · You outlived 22.2% of your peers              ║
 ╚═══════════════════════════════════════════════════════════════════════════════════════╝
```

## Features

- **Real actuarial core.** WHO life tables (2019) for the world, Russia and the
  USA, by sex, interpolated to single years, split into six cause groups with
  WHO Global Health Estimates. The engine applies your risk modifiers to each
  cause's *hazard* — exactly how a hazard ratio works.
- **106 events, 234 choices, 116 book items** across 20 chapters: roads and
  fire, smoking and salt, emergencies, the home medicine cabinet, loans and
  guarantees, legal red lines for techies, work injuries, grief, looks,
  relationships, travel.
- **Evidence card after every decision**: chapter, item, grade A/B/C, the
  book's plain-language summary, HR/OR/RR with confidence intervals, source,
  DOI and a deep link to the item in the book.
- **Autopsy report** with the decisive decision, years of life lost from the
  life table, the share of peers you outlived and your revealed hidden
  modifiers.
- **Bilingual**: Russian and English, switchable live with `L`. Language is
  detected from `$LANG`.
- **Six professions**, three colour themes (Nord, Cyberpunk, Monokai),
  willpower as a resource, conditions that last for years, chained events
  (sign a guarantee at 25, meet the bank at 27), prison terms, debt spirals.
- **Reproducible lives**: `-seed N` or `-daily` (everyone gets the same life
  today). A graveyard keeps your past runs.
- **Vim keys**, arrows, number keys; fits an 80×24 terminal.
- **One static binary**, no runtime dependencies: Arch, Kali, Parrot, Ubuntu,
  Alpine, Raspberry Pi (armv7/arm64), macOS.

## Install

**One line (Linux, macOS)** — downloads the release, verifies its SHA-256 and
installs to `~/.local/bin`:

```sh
curl -fsSL https://raw.githubusercontent.com/funstorm15/survive85/main/scripts/install.sh | sh
```

**Arch Linux (AUR)**:

```sh
yay -S survive85        # builds from source
yay -S survive85-bin    # prebuilt binary
```

**Go**:

```sh
go install github.com/funstorm15/survive85/cmd/survive85@latest
```

**Debian / Ubuntu / Kali / Parrot, Fedora, Alpine**: `.deb`, `.rpm` and `.apk`
packages are attached to every [release](https://github.com/funstorm15/survive85/releases).

**From source**:

```sh
git clone https://github.com/funstorm15/survive85 && cd survive85
make build && ./survive85
```

## Play

```sh
survive85                 # language from $LANG
survive85 -lang ru        # по-русски
survive85 -seed 42        # a reproducible life
survive85 -daily          # today's shared life
survive85 -theme cyberpunk
```

| Key | Action |
|---|---|
| `↑` `↓` / `k` `j` | move |
| `Enter` `Space` `l` | choose / confirm |
| `h` `Esc` | back |
| `1`–`4` | quick-pick a choice |
| `PgUp` `PgDn` | scroll long screens |
| `L` | Russian ⇄ English |
| `T` | cycle theme |
| `?` | help |
| `q` `Ctrl+C` | quit |
| `r` (after death) | replay the same fate |

**Stats.** *Health* (0–100) scales disease mortality relative to what is
typical at your age; at 0 you die. *Sanity* (0–100) drives the risk of
suicide; low sanity unlocks crisis events. *Wealth* buys the wise options;
debt costs 20% a year. *Willpower* (◆, 0–10) is spent on disciplined choices
and recovers every year. *Hidden modifiers* — cardiovascular, oncology,
liver/toxicology, trauma/accident and legal/prison — are revealed only in the
autopsy.

## How the mortality model works

Every year at age *a*:

1. Take the life-table probability *q(a)* for your region and sex and convert it
   to a force of mortality *μ = −ln(1 − q)*.
2. Split *μ* into six cause groups by WHO shares for your age band:
   cardiovascular, oncology, liver & poisoning, trauma, self-harm, other.
3. Multiply each cause's hazard by your modifier for it (product of every
   active factor and condition, with an overlap correction above ×2 and a clamp
   to [0.25, 25]). Disease causes are also scaled by Health relative to the
   typical Health for your age; self-harm by Sanity.
4. Die with probability *1 − exp(−Σ adjusted hazards)*; the cause is drawn in
   proportion to the adjusted hazards.

With all modifiers at 1 the game reproduces the WHO table exactly (tested).
Book numbers become factors as documented in
[docs/CONTENT_GUIDE.md](docs/CONTENT_GUIDE.md): a hazard ratio on the matching
outcome is used directly; a ratio on a narrower outcome is diluted by its
share of the cause group (the arithmetic sits in a YAML comment next to every
factor); situational effects (a helmet in a crash) become instant checks that
keep the book's ratio between choices. Absolute instant-check chances are
marked as game approximations.

**Balance** (`survive85 sim`, 4,000 lives per row):

| Life table | WHO baseline | Careful | Random | Reckless |
|---|---:|---:|---:|---:|
| World, male | 27.6% | 36.2% | 3.6% | 0.1% |
| World, female | 41.7% | 48.0% | 5.8% | 0.0% |
| Russia, male | 15.7% | 28.3% | 2.5% | 0.0% |
| Russia, female | 38.2% | 47.4% | 5.5% | 0.0% |
| USA, male | 37.5% | 43.0% | 5.0% | 0.1% |
| USA, female | 50.5% | 53.5% | 7.5% | 0.0% |

## Project layout

```
cmd/survive85/        entry point: play, sim, validate, version
internal/actuary/     life tables: q(x), e(x), cause shares
internal/events/      data model, YAML loader + validator, event registry
internal/game/        player state, risk calculator, year-by-year engine
internal/sim/         headless Monte-Carlo balance simulator
internal/i18n/        ru/en text, locale detection, money formatting
internal/tui/         Bubble Tea app, screens, Lip Gloss themes, components
data/                 embedded YAML: events, evidence, conditions, life tables
tools/genlifetables/  regenerates life tables from the WHO APIs
tools/validatedata/   validates data (optionally one chapter at a time)
scripts/              install.sh, PKGBUILD, PKGBUILD-bin
```

`make help` lists every task: `build`, `test`, `lint`, `validate`, `sim`,
`lifetables`, `snapshot`, `aur`, `install`.

## Contributing

New events are welcome — every number must come from the book, and every
event needs both languages. Read [CONTRIBUTING.md](CONTRIBUTING.md) and
[docs/CONTENT_GUIDE.md](docs/CONTENT_GUIDE.md); `go test ./data/` checks
schema, translations and references.

## Credits & licences

- **Code**: MIT © 2026 funstorm15.
- **Science**: [HowToLiveBetter](https://github.com/eternity4719/HowToLiveBetter)
  by eternity4719, CC BY 4.0 — items selected, quoted and turned into game
  events. [Online search](https://eternity4719.github.io/HowToLiveBetter/).
- **Translations**: [dlgrv/HowToLiveBetter](https://github.com/dlgrv/HowToLiveBetter)
  ([ru](https://book.dlgrv.com/ru/), [en](https://book.dlgrv.com/en/)).
- **Mortality data**: World Health Organization — Global Health Observatory
  life tables and Global Health Estimates.
- **TUI**: [Bubble Tea](https://github.com/charmbracelet/bubbletea),
  [Lip Gloss](https://github.com/charmbracelet/lipgloss),
  [Bubbles](https://github.com/charmbracelet/bubbles).

The game data is CC BY 4.0 — see [data/LICENSE.md](data/LICENSE.md) for how
to attribute it.

> survive85 is a game, not medical, legal or financial advice. If you are
> struggling, call your local emergency number or a crisis line, and talk to
> someone close to you.
