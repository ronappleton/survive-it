# Traps

Source: `internal/game/trapping.go` (`TrapCatalog`).

Total traps: **17**.

| ID | Name | Targets | Min Bushcraft | Base Catch | Setup (h) | Cond Loss | Yield (kg) | Needs Water | Requires Crafted | Requires Resources | Requires Kit | Biomes |
| --- | --- | --- | --- | --- | --- | --- | --- | --- | --- | --- | --- | --- |
| bird_noose_perch | Bird Noose Perch | bird | 1 | 20% | 0.7 | 6 | 0.1-0.5 | no | natural_twine |  |  | forest, coast, savanna, wetlands, badlands |
| crab_pot | Crab Pot | fish | 1 | 23% | 1.4 | 4 | 0.2-1.4 | yes | fish_trap_basket, heavy_cordage |  |  | coast, delta, island, wetlands |
| eel_pot | Eel Pot | fish | 2 | 27% | 1.5 | 4 | 0.2-1.6 | yes | fish_trap_basket |  |  | river, delta, swamp, wetlands, coast |
| figure4_deadfall | Figure-4 Deadfall | small_game, bird | 2 | 21% | 1.1 | 9 | 0.14-0.85 | no | trap_trigger_set |  |  | forest, boreal, mountain, savanna, desert, badlands |
| fish_weir | Fish Weir | fish | 2 | 33% | 2.2 | 5 | 0.4-2.8 | yes | fish_weir_stakes |  |  | river, delta, wetlands, coast |
| funnel_fish_basket | Funnel Fish Basket | fish | 1 | 24% | 1.2 | 4 | 0.2-1.4 | yes | fish_trap, natural_twine |  |  | delta, river, lake, swamp, coast, wetlands |
| gill_net_set | Gill Net Set | fish | 3 | 36% | 1.8 | 6 | 0.5-3.5 | yes | gill_net |  |  | coast, delta, river, lake, wetlands |
| gorge_hook_line | Gorge Hook Line | fish | 1 | 27% | 0.6 | 5 | 0.12-0.7 | yes | fish_gorge_hooks, natural_twine |  |  | river, lake, delta, coast, wetlands |
| paiute_deadfall | Paiute Deadfall | small_game | 3 | 26% | 1.3 | 8 | 0.18-1 | no | trap_trigger_set, natural_twine |  |  | forest, boreal, mountain, badlands, desert |
| paracord_twitchup | Paracord Twitch-Up | small_game | 1 | 31% | 0.65 | 7 | 0.2-1.2 | no | trap_trigger_set |  | Paracord (50 ft), Snare Wire | forest, boreal, mountain, savanna, badlands |
| peg_snare | Peg Snare | small_game | 1 | 19% | 0.5 | 7 | 0.18-0.9 | no | natural_twine |  |  | forest, boreal, savanna, badlands, tundra |
| reptile_noose | Reptile Noose | reptile | 2 | 20% | 0.9 | 7 | 0.15-1.8 | no | natural_twine |  |  | desert, dry, savanna, jungle, wetlands |
| rolling_log_deadfall | Rolling Log Deadfall | small_game, medium_game | 3 | 20% | 1.6 | 10 | 0.5-3.2 | no | deadfall_kit, wood_mallet |  |  | forest, mountain, boreal, badlands |
| snare_fence | Snare Fence | small_game | 2 | 28% | 1.4 | 7 | 0.25-1.3 | no | spring_snare_kit, heavy_cordage |  |  | savanna, badlands, forest, boreal, mountain |
| spring_snare | Spring Snare | small_game | 2 | 23% | 0.8 | 8 | 0.2-1.1 | no | trap_trigger_set, natural_twine |  |  | forest, boreal, savanna, mountain, badlands |
| trail_twitchup | Trail Twitch-Up | small_game | 2 | 25% | 0.9 | 8 | 0.2-1 | no | spring_snare_kit |  |  | forest, boreal, mountain, savanna |
| trotline | Trotline | fish | 2 | 31% | 1.2 | 4 | 0.3-2.6 | yes | trotline_set |  |  | river, lake, delta, coast |
