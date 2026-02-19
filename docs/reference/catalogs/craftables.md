# Craftables

Source: `internal/game/environment_resources.go` (`CraftableCatalog`).

Total craftables: **48**.

| ID | Name | Category | Min Bushcraft | Time (h) | Portable | Req Fire | Req Shelter | Wood (kg) | Weight (kg) | Requires Items | Requires Resources | Biomes |
| --- | --- | --- | --- | --- | --- | --- | --- | --- | --- | --- | --- | --- |
| bast_sandals | Bast Sandals | clothing | 0 | 1 | yes | no | no | 0 | 0.35 |  | bast_strip 1, yucca_fiber 1 | forest, boreal, savanna, desert, coast |
| grass_cape | Grass Cape | clothing | 0 | 1.5 | yes | no | no | 0 | 0.8 | natural_twine | dry_grass 2 | savanna, badlands, wetlands, coast, forest |
| hide_jacket | Hide Jacket | clothing | 2 | 4.6 | yes | no | no | 0 | 2 | natural_twine | rawhide_strip 2 | forest, boreal, subarctic, mountain, badlands |
| hide_moccasins | Hide Moccasins | clothing | 1 | 2.2 | yes | no | no | 0 | 0.5 | natural_twine | rawhide_strip 1 | forest, savanna, badlands, mountain, coast |
| woven_tunic | Woven Tunic | clothing | 1 | 3 | yes | no | no | 0 | 1.1 | natural_twine | nettle_fiber 2 | forest, boreal, wetlands, jungle, coast |
| heavy_cordage | Heavy Cordage | cordage | 1 | 0.8 | yes | no | no | 0 | 0.16 | natural_twine | inner_bark_fiber 1 | forest, boreal, savanna, jungle, wetlands, coast |
| natural_twine | Natural Twine | cordage | 0 | 0.45 | yes | no | no | 0 | 0.08 |  | inner_bark_fiber 1 | forest, boreal, savanna, jungle, wetlands, desert, coast |
| arrow_shafts | Arrow Shafts | general | 2 | 0 | no | no | no | 0.35 | 0 | bow_stave |  | forest, savanna, badlands, mountain, coast |
| bark_container | Bark Container | general | 1 | 0 | no | no | no | 0.2 | 0 |  | birch_bark 1 | forest, boreal, coast, mountain, jungle |
| bearing_block | Bearing Block | general | 1 | 0 | no | no | no | 0.18 | 0 |  |  | forest, coast, mountain, jungle, savanna, badlands, desert |
| bow_drill_bow | Bow Drill Bow | general | 1 | 0 | no | no | no | 0.3 | 0 |  | vine_fiber 1 | forest, coast, mountain, jungle, savanna, badlands, desert |
| bow_drill_hearth_board | Bow Drill Hearth Board | general | 1 | 0 | no | no | no | 0.25 | 0 |  |  | forest, coast, mountain, jungle, savanna, badlands, desert |
| bow_drill_spindle | Bow Drill Spindle | general | 1 | 0 | no | no | no | 0.2 | 0 |  |  | forest, coast, mountain, jungle, savanna, badlands, desert |
| bow_stave | Bow Stave | general | 2 | 0 | no | no | no | 0.9 | 0 |  | inner_bark_fiber 1 | forest, savanna, badlands, mountain, coast |
| bug_smudge | Bug Smudge Fire | general | 1 | 0 | no | yes | no | 0.5 | 0 |  | resin 1 | jungle, swamp, wetlands, tropical, island |
| carving_wedges | Carving Wedges | general | 1 | 0 | no | no | no | 0.3 | 0 |  |  | forest, mountain, coast, boreal, savanna |
| char_box | Char Box | general | 2 | 0 | no | yes | no | 0.4 | 0 |  | charcoal 1 | forest, boreal, mountain, coast, savanna, badlands |
| charcoal_bed | Charcoal Bed | general | 2 | 0 | no | yes | no | 1.4 | 0 |  |  | desert, dry, savanna, badlands |
| clay_cook_plate | Clay Cook Plate | general | 2 | 0 | no | yes | no | 0.3 | 0 |  | clay 1 | river, delta, wetlands, swamp, lake, badlands, coast |
| clay_heat_core | Clay Heat Core | general | 3 | 0 | no | yes | yes | 0.6 | 0 |  | clay 1.5 | river, delta, wetlands, swamp, lake, badlands, coast |
| clay_pot | Clay Pot | general | 2 | 0 | no | yes | no | 0.4 | 0 |  | clay 1.2 | river, delta, wetlands, swamp, lake, badlands, coast |
| tripod | Cooking Tripod | general | 0 | 0 | no | yes | no | 0.8 | 0 |  |  | forest, coast, mountain, jungle |
| digging_stick | Digging Stick | general | 0 | 0 | no | no | no | 0.35 | 0 |  |  | desert, dry, savanna, forest, jungle, badlands |
| fish_gorge_hooks | Fish Gorge Hooks | general | 2 | 0 | no | no | no | 0.18 | 0 |  | inner_bark_fiber 1 | river, lake, delta, coast, wetlands |
| fish_spear_shaft | Fish Spear Shaft | general | 1 | 0 | no | no | no | 0.4 | 0 |  |  | delta, river, lake, coast, wetlands, jungle |
| fish_trap | Fish Trap | general | 1 | 0 | no | no | no | 0.7 | 0 |  | vine_fiber 1 | delta, river, lake, swamp, coast |
| hand_drill_hearth_board | Hand Drill Hearth Board | general | 1 | 0 | no | no | no | 0.22 | 0 |  |  | forest, coast, mountain, jungle, savanna, badlands, desert |
| hand_drill_spindle | Hand Drill Spindle | general | 1 | 0 | no | no | no | 0.2 | 0 |  |  | forest, coast, mountain, jungle, savanna, badlands, desert |
| pack_frame | Pack Frame | general | 2 | 0 | no | no | no | 1.4 | 0 |  | inner_bark_fiber 1 | forest, mountain, boreal, savanna, badlands |
| pitch_glue | Pitch Glue | general | 2 | 0 | no | yes | no | 0.2 | 0 |  | resin 1, charcoal 1 | forest, boreal, mountain, savanna, jungle |
| rain_catcher | Rain Catcher | general | 1 | 0 | no | no | yes | 0.6 | 0 |  | reed_bundle 1 | jungle, wetlands, coast, island, forest |
| raised_bed | Raised Bed | general | 2 | 0 | no | no | yes | 1.1 | 0 |  | reed_bundle 1 | swamp, wetlands, jungle, forest |
| resin_torch | Resin Torch | general | 1 | 0 | no | no | no | 0.35 | 0 |  | resin 1, inner_bark_fiber 1 | forest, boreal, mountain, jungle, savanna |
| ridge_pole_kit | Ridge Pole Kit | general | 1 | 0 | no | no | no | 1.1 | 0 |  |  | forest, mountain, boreal, coast, jungle |
| shelter_lattice | Shelter Lattice | general | 1 | 0 | no | no | yes | 1.2 | 0 |  |  | forest, jungle, wetlands, swamp, coast |
| signal_beacon | Signal Beacon | general | 0 | 0 | no | yes | no | 1.3 | 0 |  |  | coast, island, mountain, badlands, savanna |
| smoke_rack | Smoke Rack | general | 1 | 0 | no | yes | no | 1.2 | 0 |  |  | forest, coast, savanna, jungle |
| snow_melt_station | Snow Melt Station | general | 1 | 0 | no | yes | no | 0.9 | 0 |  |  | arctic, subarctic, tundra, winter |
| split_basket | Split-Wood Basket | general | 2 | 0 | no | no | no | 0.9 | 0 |  | reed_bundle 1 | forest, boreal, wetlands, delta, jungle |
| tarp_stakes | Tarp Stakes | general | 0 | 0 | no | no | no | 0.25 | 0 |  |  | forest, coast, mountain, jungle, savanna, badlands |
| trap_trigger_set | Trap Trigger Set | general | 2 | 0 | no | no | no | 0.45 | 0 |  |  | forest, boreal, mountain, savanna, badlands |
| walking_staff | Walking Staff | general | 0 | 0 | no | no | no | 0.45 | 0 |  |  | forest, mountain, badlands, savanna, coast |
| windbreak | Windbreak Wall | general | 1 | 0 | no | no | yes | 1 | 0 |  |  | arctic, subarctic, mountain, badlands, coast |
| wooden_cup | Wooden Cup | general | 1 | 0 | no | no | no | 0.3 | 0 |  |  | forest, boreal, mountain, coast, savanna |
| wooden_spoon | Wooden Spoon | general | 0 | 0 | no | no | no | 0.08 | 0 |  |  | forest, boreal, mountain, coast, jungle |
| brush_raft | Brush Raft | watercraft | 1 | 3.5 | no | no | yes | 3.2 | 6.4 | heavy_cordage | reed_bundle 2 | river, lake, delta, wetlands, coast, island, jungle |
| dugout_canoe | Dugout Canoe | watercraft | 3 | 8.5 | no | yes | yes | 5.4 | 14 | heavy_cordage | charcoal 2 | river, lake, delta, coast, wetlands, forest |
| reed_coracle | Reed Coracle | watercraft | 2 | 5.2 | no | no | yes | 2 | 8.5 | pitch_glue, heavy_cordage | reed_bundle 4 | delta, wetlands, river, lake, coast |
