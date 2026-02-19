# Craftables

Source: `internal/game/environment_resources.go` (`CraftableCatalog`).

Total craftables: **100**.

| ID | Name | Category | Min Bushcraft | Time (h) | Portable | Req Fire | Req Shelter | Wood (kg) | Weight (kg) | Requires Items | Requires Resources | Biomes |
| --- | --- | --- | --- | --- | --- | --- | --- | --- | --- | --- | --- | --- |
| bark_cloak | Bark Cloak | clothing | 1 | 2.6 | yes | no | no | 0 | 1.2 | natural_twine | cedar_bark 1, bast_strip 1 | forest, boreal, coast, wetlands, jungle |
| bast_sandals | Bast Sandals | clothing | 0 | 1 | yes | no | no | 0 | 0.35 |  | bast_strip 1, yucca_fiber 1 | forest, boreal, savanna, desert, coast |
| fiber_poncho | Fiber Poncho | clothing | 1 | 1.8 | yes | no | no | 0 | 0.9 | natural_twine | hemp_fiber 1 | forest, coast, wetlands, jungle, savanna |
| grass_cape | Grass Cape | clothing | 0 | 1.5 | yes | no | no | 0 | 0.8 | natural_twine | dry_grass 2 | savanna, badlands, wetlands, coast, forest |
| hide_jacket | Hide Jacket | clothing | 2 | 4.6 | yes | no | no | 0 | 2 | natural_twine | rawhide_strip 2 | forest, boreal, subarctic, mountain, badlands |
| hide_moccasins | Hide Moccasins | clothing | 1 | 2.2 | yes | no | no | 0 | 0.5 | natural_twine | rawhide_strip 1 | forest, savanna, badlands, mountain, coast |
| hide_trousers | Hide Trousers | clothing | 2 | 3.8 | yes | no | no | 0 | 1.5 | bone_needle, natural_twine | rawhide_strip 2 | forest, boreal, subarctic, mountain, badlands |
| reed_sandals | Reed Sandals | clothing | 0 | 0.9 | yes | no | no | 0 | 0.3 | natural_twine | reed_bundle 1 | wetlands, swamp, delta, coast, jungle |
| woven_tunic | Woven Tunic | clothing | 1 | 3 | yes | no | no | 0 | 1.1 | natural_twine | nettle_fiber 2 | forest, boreal, wetlands, jungle, coast |
| heavy_cordage | Heavy Cordage | cordage | 1 | 0.8 | yes | no | no | 0 | 0.16 | natural_twine | inner_bark_fiber 1 | forest, boreal, savanna, jungle, wetlands, coast |
| natural_twine | Natural Twine | cordage | 0 | 0.45 | yes | no | no | 0 | 0.08 |  | inner_bark_fiber 1 | forest, boreal, savanna, jungle, wetlands, desert, coast |
| char_cloth | Char Cloth | fire | 1 | 0.7 | yes | yes | no | 0 | 0.04 |  | flax_fiber 1, charcoal 1 | forest, boreal, coast, savanna, jungle |
| ember_pot | Ember Pot | fire | 2 | 2.2 | yes | yes | no | 0 | 1.2 |  | clay 1 | forest, boreal, mountain, coast, river |
| fish_trap_basket | Fish Trap Basket | fishing | 1 | 1.8 | yes | no | no | 0 | 1.2 | natural_twine | reed_bundle 2 | delta, river, lake, swamp, coast, wetlands |
| fish_weir_stakes | Fish Weir Stakes | fishing | 2 | 2.6 | no | no | no | 0 | 2 | heavy_cordage | willow_withy 2 | river, delta, wetlands, coast |
| gill_net | Gill Net | fishing | 3 | 4 | yes | no | no | 0 | 2.2 | heavy_cordage | hemp_fiber 2 | coast, delta, lake, river, wetlands |
| trotline_set | Trotline Set | fishing | 2 | 2 | yes | no | no | 0 | 0.9 | fish_gorge_hooks, heavy_cordage |  | river, lake, delta, coast |
| clay_oven | Clay Oven | food | 2 | 4.5 | no | yes | yes | 0.8 | 10 |  | clay 2.2, gravel 1 | river, delta, wetlands, swamp, lake, coast |
| drying_rack | Drying Rack | food | 1 | 2.8 | no | no | yes | 1.4 | 4.5 | heavy_cordage |  | forest, coast, savanna, jungle, mountain |
| smoking_rack | Smoking Rack | food | 2 | 3.4 | no | yes | yes | 1.8 | 5 | drying_rack |  | forest, coast, savanna, jungle, wetlands |
| stone_oven | Stone Oven | food | 2 | 5 | no | yes | yes | 1 | 14 |  | stone_cobble 3, mud 1 | mountain, badlands, river, coast, forest |
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
| atlatl | Atlatl Thrower | hunting | 2 | 2 | yes | no | no | 0.6 | 0.7 | fire_hardened_spear, natural_twine |  | forest, mountain, savanna, badlands, coast |
| bone_arrow_bundle | Bone Arrow Bundle | hunting | 2 | 1.7 | yes | no | no | 0 | 0.6 | short_bow, natural_twine | shell_fragment 2 | forest, boreal, wetlands, coast, jungle |
| fire_hardened_spear | Fire-Hardened Spear | hunting | 1 | 1.4 | yes | yes | no | 0.8 | 1.2 |  |  | forest, coast, river, jungle, savanna, badlands |
| long_bow | Long Bow | hunting | 3 | 4.1 | yes | no | no | 1.1 | 1.3 | short_bow, heavy_cordage |  | forest, boreal, mountain, coast |
| short_bow | Short Bow | hunting | 2 | 3.2 | yes | no | no | 0.8 | 0.9 | heavy_cordage |  | forest, coast, savanna, badlands, mountain |
| stone_arrow_bundle | Stone Arrow Bundle | hunting | 2 | 1.8 | yes | no | no | 0 | 0.7 | short_bow, natural_twine | stone_flake 2 | forest, mountain, badlands, coast, savanna |
| bough_mattress | Bough Mattress | shelter_upgrade | 1 | 1.1 | no | no | yes | 0 | 2 |  | spruce_bough 2 | forest, boreal, wetlands, mountain |
| camouflage_screen | Camouflage Screen | shelter_upgrade | 1 | 1.3 | no | no | yes | 0 | 0.8 | natural_twine | dry_leaf_litter 1 | forest, savanna, wetlands, jungle, mountain |
| cold_air_trench | Cold-Air Trench | shelter_upgrade | 1 | 1 | no | no | yes | 0 | 0 | digging_stick |  | arctic, subarctic, tundra, boreal, mountain |
| door_latch | Door Latch | shelter_upgrade | 1 | 0.8 | no | no | yes | 0.4 | 0.3 | bone_awl |  | forest, boreal, mountain, coast, badlands |
| drainage_ditch | Drainage Ditch | shelter_upgrade | 1 | 1 | no | no | yes | 0 | 0 | digging_stick | gravel 1 | forest, wetlands, swamp, mountain, coast |
| groundsheet | Groundsheet | shelter_upgrade | 1 | 1.4 | no | no | yes | 0 | 1 |  | bark_pitch 1, reed_bundle 1 | forest, coast, wetlands, savanna, jungle |
| insulated_bedding | Insulated Bedding | shelter_upgrade | 2 | 2.2 | no | no | yes | 0 | 1.6 |  | dry_moss 2, rawhide_strip 1 | forest, boreal, tundra, mountain, wetlands |
| insulated_wall_lining | Insulated Wall Lining | shelter_upgrade | 2 | 2.1 | no | no | yes | 0 | 1.7 |  | dry_moss 1, thatch_bundle 1 | forest, boreal, subarctic, mountain, wetlands |
| raised_sleeping_platform | Raised Sleeping Platform | shelter_upgrade | 2 | 2.6 | no | no | yes | 1.6 | 5.5 |  | spruce_bough 1 | forest, wetlands, swamp, jungle, coast |
| reflective_fire_wall | Reflective Fire Wall | shelter_upgrade | 2 | 2.1 | no | yes | yes | 1.2 | 2.5 |  | stone_cobble 1 | forest, mountain, boreal, badlands |
| smoke_hole_baffle | Smoke-Hole Baffle | shelter_upgrade | 2 | 1.8 | no | no | yes | 0.7 | 1.2 |  | thatch_bundle 1 | forest, boreal, mountain, tundra, wetlands |
| stone_hearth | Stone Hearth | shelter_upgrade | 2 | 2.8 | no | yes | yes | 0 | 8 |  | stone_cobble 2, mud 0.8 | forest, mountain, badlands, coast, boreal |
| storage_shelves | Storage Shelves | shelter_upgrade | 1 | 1.9 | no | no | yes | 1.1 | 3 |  |  | forest, coast, mountain, jungle, boreal |
| storm_flap | Storm Flap | shelter_upgrade | 1 | 1.2 | no | no | yes | 0 | 0.7 | natural_twine | thatch_bundle 1 | forest, coast, wetlands, mountain, tundra |
| drying_box | Drying Box | storage | 1 | 2 | no | no | yes | 1.2 | 3.4 | drying_rack |  | forest, coast, savanna, jungle, badlands |
| elevated_food_cache | Elevated Food Cache | storage | 2 | 2.7 | no | no | yes | 1.7 | 4.2 | heavy_cordage |  | forest, boreal, wetlands, tundra, mountain |
| underground_cold_pit | Underground Cold Pit | storage | 2 | 2.5 | no | no | yes | 0 | 0 | digging_stick | gravel 1, clay 0.8 | forest, boreal, mountain, badlands, tundra |
| hunting_blind | Hunting Blind | structures | 1 | 2.1 | no | no | yes | 1.3 | 3.5 |  | thatch_bundle 1 | forest, savanna, wetlands, badlands, mountain |
| lookout_platform | Lookout Platform | structures | 2 | 3.8 | no | no | yes | 2.4 | 9 | wedge_set |  | forest, savanna, wetlands, badlands, coast |
| perimeter_deadfall_deterrent | Perimeter Deadfall Deterrent | structures | 2 | 2.2 | no | no | yes | 1.6 | 4.8 | deadfall_kit |  | forest, boreal, mountain, badlands, savanna |
| raised_storage_platform | Raised Storage Platform | structures | 2 | 2.9 | no | no | yes | 2 | 7 | heavy_cordage |  | forest, wetlands, swamp, delta, jungle |
| bone_awl | Bone Awl | tools | 1 | 0.7 | yes | no | no | 0 | 0.08 |  | shell_fragment 1 | forest, boreal, savanna, jungle, coast |
| bone_needle | Bone Needle | tools | 1 | 0.6 | yes | no | no | 0 | 0.03 |  | shell_fragment 1 | forest, boreal, savanna, jungle, coast |
| stone_adze | Stone Adze | tools | 2 | 2.4 | yes | no | no | 0 | 1.1 | heavy_cordage | stone_cobble 1 | forest, mountain, badlands, river, coast |
| wedge_set | Wedge Set | tools | 1 | 1 | yes | no | no | 0.6 | 0.4 | wood_mallet |  | forest, mountain, boreal, coast, badlands |
| wood_mallet | Wood Mallet | tools | 1 | 1.2 | yes | no | no | 0.9 | 0.8 |  |  | forest, mountain, boreal, savanna |
| hand_sled | Hand Sled | transport | 2 | 3.2 | no | no | no | 2.2 | 7.5 | wedge_set, heavy_cordage |  | tundra, arctic, subarctic, boreal, mountain |
| skis | Improvised Skis | transport | 3 | 4.2 | yes | no | no | 0 | 2.8 | wood_mallet, wedge_set, heavy_cordage |  | tundra, arctic, subarctic, boreal, mountain |
| snowshoes | Snowshoes | transport | 2 | 3.3 | yes | no | no | 0 | 2.1 | heavy_cordage | willow_withy 1 | tundra, arctic, subarctic, boreal, mountain |
| travois | Travois | transport | 1 | 2.4 | no | no | no | 1.8 | 6.2 | heavy_cordage |  | savanna, badlands, forest, mountain, desert |
| deadfall_kit | Deadfall Kit | trapping | 2 | 1.4 | yes | no | no | 0 | 0.7 | trap_trigger_set, wedge_set |  | forest, boreal, mountain, badlands, desert |
| spring_snare_kit | Spring Snare Kit | trapping | 2 | 1.6 | yes | no | no | 0 | 0.5 | trap_trigger_set, heavy_cordage |  | forest, boreal, savanna, mountain, badlands |
| brush_raft | Brush Raft | watercraft | 1 | 3.5 | no | no | yes | 3.2 | 6.4 | heavy_cordage | reed_bundle 2 | river, lake, delta, wetlands, coast, island, jungle |
| dugout_canoe | Dugout Canoe | watercraft | 3 | 8.5 | no | yes | yes | 5.4 | 14 | heavy_cordage | charcoal 2 | river, lake, delta, coast, wetlands, forest |
| reed_coracle | Reed Coracle | watercraft | 2 | 5.2 | no | no | yes | 2 | 8.5 | pitch_glue, heavy_cordage | reed_bundle 4 | delta, wetlands, river, lake, coast |
