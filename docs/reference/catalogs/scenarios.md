# Scenarios

Source: `internal/game/scenarios_builtin.go` (`BuiltInScenarios`).

Total built-in scenarios: **29**.

| ID | Name | Modes | Days | Location | Profile ID | BBox | Biome | Map (cells) | Default Season Set | Season Phases |
| --- | --- | --- | --- | --- | --- | --- | --- | --- | --- | --- |
| naa_alaska | Alaska Cold Region (NAA) | naked_and_afraid | 21 | North America | alaska_cold_region | -153.00, 63.00, -147.00, 67.00 | tundra | 100x100 | winter_default | winter_default[winter:end] |
| arctic | Arctic | alone | 365 | Wilderness |  |  | subarctic | 36x36 | winter_default | winter_default[winter:end] |
| chilko_lake_bc | Chilko Lake (British Columbia) | alone | 365 | Wilderness | chilko_lake | -125.50, 50.80, -123.40, 52.50 | mountain_forest | 36x36 | alone_default | alone_default[autumn:14,winter:end] |
| naa_colombia | Colombia Jungle (NAA) | naked_and_afraid | 21 | South America | colombia_jungle | -75.50, 0.50, -72.00, 3.50 | tropical_jungle | 100x100 | wet_default | wet_default[wet:end] |
| naa_costa_rica | Costa Rica Jungle (NAA) | naked_and_afraid | 21 | South America | costa_rica_jungle | -85.80, 9.00, -82.60, 11.20 | tropical_jungle | 100x100 | wet_default | wet_default[wet:end] |
| naa_florida_everglades | Florida Everglades (NAA) | naked_and_afraid | 21 | Wilderness | florida_everglades | -81.70, 24.80, -80.00, 26.50 | wetlands | 100x100 | wet_default | wet_default[wet:end] |
| great_slave_lake_100 | Great Slave Lake (Canada) - 100 Days | alone | 100 | North America | great_slave_lake | -117.50, 60.50, -108.00, 63.20 | subarctic_lake | 36x36 | winter_default | winter_default[winter:end] |
| great_slave_lake_365 | Great Slave Lake (Canada) - 365 Days | alone | 365 | North America | great_slave_lake | -117.50, 60.50, -108.00, 63.20 | subarctic_lake | 36x36 | winter_default | winter_default[winter:end] |
| jungle | Jungle | naked_and_afraid | 21 | Wilderness |  |  | tropical_jungle | 100x100 | wet_default | wet_default[wet:end] |
| mongolia_khentii | Khentii Mountains (Mongolia) | alone | 365 | Asia-Pacific | khentii_mountains | 107.00, 47.00, 111.80, 50.20 | montane_steppe | 36x36 | dry_default | dry_default[dry:end] |
| labrador_coast | Labrador Coast (Canada) | alone | 365 | North America | labrador_coast | -61.80, 52.50, -55.00, 57.00 | boreal_coast | 36x36 | winter_default | winter_default[winter:end] |
| naa_louisiana_swamp | Louisiana Swamp (NAA) | naked_and_afraid | 21 | North America | louisiana_swamp | -92.20, 29.00, -89.00, 31.20 | swamp | 100x100 | wet_default | wet_default[wet:end] |
| mackenzie_delta | Mackenzie River Delta (NWT) | alone | 365 | Wilderness | mackenzie_delta | -137.80, 67.10, -132.00, 70.00 | arctic_delta | 36x36 | winter_default | winter_default[winter:end] |
| naaxl_colombia_40 | NAA XL Colombia (40) | naked_and_afraid_xl | 40 | South America | colombia_jungle | -75.80, 1.00, -72.50, 4.00 | badlands_jungle_edge | 125x125 | dry_default | dry_default[dry:end] |
| naaxl_ecuador_40 | NAA XL Ecuador (40) | naked_and_afraid_xl | 40 | South America | ecuador_jungle | -78.90, -2.40, -76.00, 0.70 | tropical_jungle | 125x125 | wet_default | wet_default[wet:end] |
| naaxl_montana_frozen_14 | NAA XL Frozen Montana (14) | naked_and_afraid_xl | 14 | Wilderness | montana_frozen | -113.00, 45.60, -109.70, 48.00 | cold_mountain | 125x125 | winter_default | winter_default[winter:end] |
| naaxl_louisiana_60 | NAA XL Louisiana (60) | naked_and_afraid_xl | 60 | North America | louisiana_swamp | -92.20, 29.00, -89.00, 31.20 | swamp | 125x125 | wet_default | wet_default[wet:end] |
| naaxl_nicaragua_40 | NAA XL Nicaragua (40) | naked_and_afraid_xl | 40 | South America | nicaragua_jungle | -86.80, 10.80, -83.50, 13.20 | tropical_jungle | 125x125 | wet_default | wet_default[wet:end] |
| naaxl_philippines_40 | NAA XL Philippines (40) | naked_and_afraid_xl | 40 | Asia-Pacific | philippines_island | 120.30, 13.00, 123.50, 15.80 | tropical_island | 125x125 | wet_default | wet_default[wet:end] |
| naaxl_south_africa_40 | NAA XL South Africa (40) | naked_and_afraid_xl | 40 | Africa | south_africa_savanna | 23.00, -26.70, 27.00, -23.00 | savanna | 125x125 | dry_default | dry_default[dry:end] |
| naa_namibia | Namib Desert (NAA) | naked_and_afraid | 21 | Africa | namib_desert | 14.00, -25.50, 16.80, -22.00 | desert | 100x100 | dry_default | dry_default[dry:end] |
| naa_nicaragua | Nicaragua Jungle (NAA) | naked_and_afraid | 21 | South America | nicaragua_jungle | -86.80, 10.80, -83.50, 13.20 | tropical_jungle | 100x100 | wet_default | wet_default[wet:end] |
| naa_panama | Panama Survival (NAA) | naked_and_afraid | 21 | South America | panama_darien | -79.00, 7.20, -77.20, 9.20 | tropical_jungle | 100x100 | wet_default | wet_default[wet:end] |
| patagonia_argentina | Patagonia (Argentina) | alone | 365 | South America | patagonia_argentina | -72.50, -51.50, -67.00, -46.00 | cold_steppe | 36x36 | alone_default | alone_default[autumn:14,winter:end] |
| naa_philippines | Philippines Island (NAA) | naked_and_afraid | 21 | Asia-Pacific | philippines_island | 120.30, 13.00, 123.50, 15.80 | tropical_island | 100x100 | wet_default | wet_default[wet:end] |
| reindeer_lake | Reindeer Lake (Saskatchewan) | alone | 365 | North America | reindeer_lake | -104.70, 55.00, -101.00, 58.80 | boreal_lake | 36x36 | winter_default | winter_default[winter:end] |
| naa_tanzania | Tanzania Savanna (NAA) | naked_and_afraid | 21 | Africa | tanzania_savanna | 33.00, -3.50, 35.60, -1.20 | savanna | 100x100 | dry_default | dry_default[dry:end] |
| vancouver_island | Vancouver Island | alone | 365 | North America | vancouver_island | -128.50, 48.20, -123.00, 50.90 | temperate_rainforest | 36x36 | alone_default | alone_default[autumn:14,winter:end] |
| naa_mexico_yucatan | Yucatan (NAA) | naked_and_afraid | 21 | South America | yucatan_dry_forest | -90.80, 18.40, -87.20, 21.80 | tropical_dry_forest | 100x100 | dry_default | dry_default[dry:end] |
