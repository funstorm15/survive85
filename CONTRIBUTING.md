# Contributing to survive85

Thanks for helping! The most valuable contributions are **new events** built
on the book [HowToLiveBetter](https://github.com/eternity4719/HowToLiveBetter).

## Ground rules

1. **No invented science.** Every number, grade, DOI and quote comes from the
   book item you cite (`data/evidence/chNN.yaml`). Game-only numbers (an
   absolute chance in an instant check, a price in rubles) must be commented
   as `# game approximation`.
2. **Both languages.** Every player-visible string needs `ru` and `en`, each
   written naturally — not a word-for-word translation.
3. **Show your arithmetic.** When a book ratio becomes a factor, write the
   calculation in a YAML comment next to it.

The full schema, the risk keys and the rules for turning HR/OR/RR into
factors are in [docs/CONTENT_GUIDE.md](docs/CONTENT_GUIDE.md).

## Workflow

```sh
make test                                   # unit tests, including data validation
go run ./tools/validatedata -only 02,ch02   # validate one chapter in isolation
make sim                                    # check the balance didn't break
make run ARGS="-lang ru -seed 42"           # play it
make lint                                   # golangci-lint v2
```

A change to the engine should keep `TestAnnualMortalityBaselineMatchesLifeTable`
green: with every modifier at 1 the game must reproduce the WHO tables.

## Pull requests

- One topic per PR; describe which book items you used.
- Keep `go test ./...` and `golangci-lint run` green (CI runs both).
- Code is MIT; data you add is released under CC BY 4.0 with the rest of
  `data/`.
