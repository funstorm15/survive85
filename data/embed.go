// Package data embeds the game content: events, evidence from the book
// "HowToLiveBetter" (CC BY 4.0), conditions, professions and WHO life tables.
package data

import "embed"

// FS holds every data file, rooted at this directory.
//
//go:embed events/*.yaml evidence/*.yaml conditions/*.yaml professions.yaml causes.yaml lifetables/*.yaml
var FS embed.FS

// LifeTables is the path of the life-table file inside FS.
const LifeTables = "lifetables/who2019.yaml"
