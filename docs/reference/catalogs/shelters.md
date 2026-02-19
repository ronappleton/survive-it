# Shelters

Source: `internal/game/environment_resources.go` (`ShelterCatalog`).

Total shelters: **16**.

| ID | Name | Storage (kg) | Insulation | Rain | Wind | Insect | Dryness | Predator | Comfort | Stealth | Durability Loss/Day | Build Cost (E/H2O) | Morale Bonus | Stages | Sleep Shelter | Upgrade Components | Biomes |
| --- | --- | --- | --- | --- | --- | --- | --- | --- | --- | --- | --- | --- | --- | --- | --- | --- | --- |
| a_frame | A-Frame Shelter | 30 | 4 | 4 | 4 | 2 | 4 | 2 | 3 | 3 | 5 | 5/2 | 3 | 3 | yes | insulated_wall_lining, storm_flap, storage_shelves, reflective_fire_wall | forest, mountain, boreal, coast, wetlands |
| bamboo_hut | Bamboo Hut | 40 | 3 | 4 | 3 | 4 | 4 | 3 | 4 | 2 | 4 | 6/3 | 4 | 0 | yes |  | jungle, tropical, wetlands, island |
| debris_hut | Debris Hut | 34 | 5 | 4 | 4 | 2 | 4 | 2 | 3 | 4 | 5 | 6/3 | 3 | 3 | yes | bough_mattress, storm_flap, drainage_ditch, reflective_fire_wall | forest, boreal, subarctic, mountain |
| desert_shade | Desert Shade | 18 | 1 | 1 | 2 | 2 | 1 | 1 | 2 | 2 | 4 | 3/2 | 2 | 0 | yes | storm_flap, reflective_fire_wall | desert, dry, savanna, badlands |
| earth_sheltered_dugout | Earth-Sheltered Dugout | 52 | 6 | 4 | 6 | 3 | 4 | 4 | 4 | 5 | 3 | 8/3 | 4 | 3 | yes | stone_hearth, smoke_hole_baffle, drainage_ditch, storage_shelves | forest, mountain, badlands, tundra, boreal |
| elevated_cache_pod | Elevated Cache Pod | 26 | 0 | 2 | 2 | 2 | 3 | 5 | 0 | 3 | 4 | 3/1 | 2 | 0 | no | elevated_food_cache | forest, wetlands, swamp, tundra, boreal |
| hunting_blind | Hunting Blind | 12 | 0 | 1 | 2 | 1 | 1 | 0 | 1 | 6 | 6 | 3/1 | 1 | 2 | no | camouflage_screen, lookout_platform | forest, savanna, wetlands, badlands, mountain |
| lean_to | Lean-to | 24 | 3 | 3 | 3 | 1 | 2 | 1 | 2 | 2 | 6 | 4/2 | 2 | 3 | yes | storm_flap, drainage_ditch, raised_sleeping_platform | forest, mountain, boreal, coast |
| log_cabin | Log Cabin | 72 | 7 | 7 | 7 | 5 | 6 | 6 | 6 | 2 | 2 | 10/4 | 5 | 6 | yes | stone_hearth, smoke_hole_baffle, storage_shelves, elevated_food_cache, door_latch | forest, boreal, mountain, subarctic |
| quinzee | Quinzee | 30 | 7 | 3 | 7 | 3 | 3 | 3 | 3 | 4 | 6 | 8/3 | 3 | 3 | yes | cold_air_trench, bough_mattress | arctic, subarctic, tundra, winter, boreal |
| raised_platform_shelter | Raised Platform Shelter | 40 | 3 | 4 | 3 | 6 | 6 | 4 | 4 | 2 | 4 | 7/3 | 4 | 3 | yes | bough_mattress, storm_flap, elevated_food_cache, drainage_ditch | swamp, wetlands, delta, jungle, coast |
| rock_overhang | Rock Overhang | 24 | 3 | 2 | 4 | 1 | 2 | 2 | 2 | 3 | 3 | 2/1 | 2 | 0 | yes |  | mountain, badlands, desert, coast |
| snow_cave | Snow Cave | 26 | 6 | 3 | 6 | 3 | 3 | 3 | 3 | 4 | 7 | 8/3 | 2 | 3 | yes | bough_mattress, cold_air_trench | arctic, subarctic, tundra, winter |
| swamp_platform | Swamp Platform | 28 | 2 | 2 | 2 | 5 | 5 | 3 | 3 | 2 | 5 | 6/3 | 3 | 0 | yes |  | swamp, wetlands, delta, jungle |
| tarp_a_frame | Tarp A-Frame | 30 | 3 | 4 | 3 | 2 | 4 | 2 | 3 | 2 | 4 | 3/2 | 3 | 0 | yes |  | forest, coast, wetlands, jungle, mountain |
| wattle_daub_hut | Wattle & Daub Hut | 44 | 5 | 5 | 5 | 3 | 5 | 3 | 4 | 2 | 3 | 7/3 | 4 | 3 | yes | storm_flap, stone_hearth, smoke_hole_baffle, storage_shelves | wetlands, swamp, river, forest, delta |
