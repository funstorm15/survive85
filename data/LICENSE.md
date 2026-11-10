# Data licence and attribution

The game **code** is MIT-licensed (see `../LICENSE`). The **data** in this
directory is licensed under the
[Creative Commons Attribution 4.0 International licence (CC BY 4.0)](https://creativecommons.org/licenses/by/4.0/),
because it is derived from CC BY 4.0 material.

## Sources

* **"HowToLiveBetter" (高性价比人生指南)** by eternity4719 —
  <https://github.com/eternity4719/HowToLiveBetter>, CC BY 4.0.
  Used in `evidence/`, `events/` and `conditions/`.
  **Changes were made:** book items were selected, quoted (titles, "in plain
  terms" and "benefit" lines, evidence grades, risk ratios, sources and DOIs)
  and turned into game events, choices and risk modifiers. Game-specific
  numbers (absolute chances of instant checks, durations, money amounts) are
  marked in the YAML as game approximations and are not claims of the book.
* **Russian and English translations** — community translation
  <https://github.com/dlgrv/HowToLiveBetter> (released under the Unlicense);
  the `plain`, `benefit` and `title` quotes in `evidence/` are taken from it.
  Items newer than the translation were translated from the Chinese original.
* **Life tables** (`lifetables/who2019.yaml`) — derived from the World Health
  Organization Global Health Observatory life tables (indicators
  LIFE_0000000030 and LIFE_0000000035) and WHO Global Health Estimates 2021
  (deaths by cause, age and sex), year 2019. Single-year probabilities were
  interpolated by `tools/genlifetables`. WHO is not responsible for this
  derivative use.

## How to attribute when you reuse the data

> Game data from survive85 (https://github.com/funstorm15/survive85), derived
> from "HowToLiveBetter" by eternity4719 (CC BY 4.0) and WHO life tables.

This game is not medical, legal or financial advice.
