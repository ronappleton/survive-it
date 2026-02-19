# Traps

Source: `internal/game/trapping.go` (`TrapCatalog`).

Total traps: **8**.

| ID | Name | Targets | Min Bushcraft | Base Catch | Setup (h) | Cond Loss | Yield (kg) | Needs Water | Requires Crafted | Requires Resources | Requires Kit | Biomes |
| --- | --- | --- | --- | --- | --- | --- | --- | --- | --- | --- | --- | --- |
| bird_noose_perch | Bird Noose Perch | bird | 1 | 20% | 0.7 | 6 | 0.1-0.5 | no | natural_twine |  |  | forest, coast, savanna, wetlands, badlands |
| figure4_deadfall | Figure-4 Deadfall | small_game, bird | 2 | 21% | 1.1 | 9 | 0.14-0.85 | no | trap_trigger_set |  |  | forest, boreal, mountain, savanna, desert, badlands |
| funnel_fish_basket | Funnel Fish Basket | fish | 1 | 24% | 1.2 | 4 | 0.2-1.4 | yes | fish_trap, natural_twine |  |  | delta, river, lake, swamp, coast, wetlands |
| gorge_hook_line | Gorge Hook Line | fish | 1 | 27% | 0.6 | 5 | 0.12-0.7 | yes | fish_gorge_hooks, natural_twine |  |  | river, lake, delta, coast, wetlands |
| paiute_deadfall | Paiute Deadfall | small_game | 3 | 26% | 1.3 | 8 | 0.18-1 | no | trap_trigger_set, natural_twine |  |  | forest, boreal, mountain, badlands, desert |
| paracord_twitchup | Paracord Twitch-Up | small_game | 1 | 31% | 0.65 | 7 | 0.2-1.2 | no | trap_trigger_set |  | Paracord (50 ft), Snare Wire | forest, boreal, mountain, savanna, badlands |
| peg_snare | Peg Snare | small_game | 1 | 19% | 0.5 | 7 | 0.18-0.9 | no | natural_twine |  |  | forest, boreal, savanna, badlands, tundra |
| spring_snare | Spring Snare | small_game | 2 | 23% | 0.8 | 8 | 0.2-1.1 | no | trap_trigger_set, natural_twine |  |  | forest, boreal, savanna, mountain, badlands |
