# Trees

Source: `internal/game/environment_resources.go` (`TreeCatalog`).

Total trees: **42**.

| ID | Name | Wood Type | Biome Tags | Gather (kg) | Heat | Burn | Spark | Hardness | Structural | Resin | Rot Resist | Smoke | Bark Resource | Bark Uses | Tags |
| --- | --- | --- | --- | --- | --- | --- | --- | --- | --- | --- | --- | --- | --- | --- | --- |
| acacia | Acacia | hardwood | savanna, badlands, dry | 0.7-3.5 | 1.08 | 1.02 | 2 | 5 | 4 | 0.18 | 4 | 0.70 | bast_strip | cordage, tannin | hardwood, dryland |
| alder | Alder | hardwood | coast, river, wetlands, forest | 0.7-3.8 | 0.95 | 0.90 | 2 | 3 | 3 | 0.12 | 3 | 0.90 | bast_strip | smoke curing, dye | wetland |
| ash | Ash | hardwood | forest, river, mountain, temperate | 0.8-4.7 | 1.12 | 1.06 | 2 | 4 | 5 | 0.12 | 4 | 0.65 | bast_strip | tool handles, splints | hardwood |
| aspen | Aspen | softwood | boreal, subarctic, forest, mountain | 0.6-3.3 | 0.80 | 0.76 | 3 | 2 | 2 | 0.08 | 2 | 0.85 | inner_bark_fiber | tinder, inner bark | boreal |
| bamboo | Bamboo | bamboo | jungle, tropical, wetlands | 0.9-5.5 | 0.78 | 0.65 | 4 | 3 | 4 | 0.05 | 3 | 0.75 | bamboo_strip | splints, containers | tropical |
| baobab | Baobab | hardwood | savanna, dry, badlands | 1.0-5.8 | 1.18 | 1.16 | 2 | 4 | 5 | 0.10 | 4 | 0.80 | bast_strip | fiber, rope | hardwood, dryland |
| beech | Beech | hardwood | forest, mountain, temperate, coast | 0.8-4.8 | 1.12 | 1.09 | 2 | 4 | 4 | 0.08 | 3 | 0.65 | bast_strip | bark strips, containers | hardwood |
| birch | Birch | hardwood | boreal, forest, lake, subarctic | 0.7-3.8 | 1.00 | 1.05 | 3 | 4 | 4 | 0.35 | 3 | 0.70 | birch_bark | container, tinder | hardwood |
| black_spruce | Black Spruce | resinous | boreal, subarctic, swamp, tundra | 0.8-4.0 | 0.93 | 0.86 | 4 | 3 | 3 | 0.78 | 4 | 1.05 | spruce_root | sewing root, pitch | conifer, wetland |
| casuarina | Casuarina | hardwood | coast, island, dry, savanna | 0.8-4.2 | 1.13 | 1.07 | 2 | 4 | 4 | 0.12 | 4 | 0.75 | bast_strip | posts, stakes | coastal |
| cedar | Cedar | softwood | coast, temperate_rainforest, vancouver, forest | 0.8-4.2 | 0.85 | 0.80 | 3 | 2 | 3 | 0.50 | 5 | 0.80 | cedar_bark | cordage, roofing | conifer |
| cottonwood | Cottonwood | softwood | river, delta, wetlands, coast | 0.7-3.5 | 0.76 | 0.72 | 3 | 2 | 2 | 0.10 | 2 | 0.90 | inner_bark_fiber | cordage, kindling | wetland |
| cypress | Cypress | softwood | swamp, wetlands, delta, coast | 0.8-4.1 | 0.90 | 0.86 | 3 | 3 | 4 | 0.40 | 5 | 0.85 | bast_strip | planking, cordage | wetland |
| driftwood | Driftwood | driftwood | coast, delta, island, lake | 0.4-2.4 | 0.62 | 0.54 | 3 | 3 | 2 | 0.03 | 3 | 1.30 |  | rafts, platforms | coastal |
| elm | Elm | hardwood | forest, river, temperate, coast | 0.8-4.6 | 1.07 | 1.04 | 2 | 4 | 5 | 0.10 | 4 | 0.70 | bast_strip | cordage, basket frame | hardwood |
| eucalyptus | Eucalyptus | resinous | savanna, dry, coast, badlands | 0.8-4.3 | 1.00 | 0.92 | 4 | 3 | 4 | 0.55 | 4 | 1.15 | bast_strip | medicine leaves, fuel | dryland |
| fir | Fir | softwood | forest, mountain, boreal | 0.8-4.0 | 0.86 | 0.82 | 3 | 2 | 3 | 0.55 | 3 | 1.00 | inner_bark_fiber | tinder, bedding | conifer |
| hemlock | Hemlock | softwood | forest, boreal, coast, mountain | 0.7-3.9 | 0.84 | 0.78 | 3 | 2 | 3 | 0.45 | 3 | 0.95 | inner_bark_fiber | tinder, fiber | conifer |
| hickory | Hickory | hardwood | forest, mountain, temperate, badlands | 0.9-5.1 | 1.20 | 1.17 | 2 | 5 | 5 | 0.10 | 4 | 0.65 | bast_strip | tool hafts, smoke curing | hardwood |
| ironwood | Ironwood | hardwood | island, coast, tropical, dry | 1.0-5.6 | 1.24 | 1.20 | 1 | 5 | 5 | 0.06 | 5 | 0.60 | bast_strip | mallets, wedges | hardwood, tropical |
| juniper | Juniper | resinous | desert, dry, mountain, badlands | 0.6-3.0 | 0.95 | 0.88 | 4 | 3 | 3 | 0.74 | 4 | 1.10 | inner_bark_fiber | tinder, smoke cure | conifer, dryland |
| kapok | Kapok | softwood | jungle, wetlands, tropical, delta | 0.7-3.6 | 0.78 | 0.73 | 3 | 2 | 2 | 0.06 | 2 | 0.85 | inner_bark_fiber | fiber floss, cordage | tropical |
| larch | Larch | resinous | boreal, subarctic, mountain, forest | 0.8-4.3 | 0.98 | 0.92 | 4 | 3 | 4 | 0.70 | 4 | 0.90 | inner_bark_fiber | cordage, tinder | conifer |
| mahogany | Mahogany | hardwood | jungle, tropical, island, coast | 0.9-5.1 | 1.09 | 1.03 | 2 | 4 | 5 | 0.10 | 5 | 0.70 | bast_strip | frames, paddles | tropical, hardwood |
| mangrove | Mangrove | hardwood | delta, coast, swamp, island | 0.8-3.8 | 1.12 | 1.06 | 2 | 4 | 4 | 0.20 | 5 | 0.85 | bast_strip | lashing, waterproofing | wetland, tropical |
| maple | Maple | hardwood | forest, mountain, temperate | 0.8-4.4 | 1.10 | 1.10 | 2 | 4 | 4 | 0.18 | 4 | 0.65 | inner_bark_fiber | fiber, container | hardwood |
| mesquite | Mesquite | hardwood | desert, dry, badlands | 0.8-4.4 | 1.20 | 1.20 | 2 | 5 | 4 | 0.16 | 5 | 0.60 | bast_strip | fiber, smoke cure | hardwood, dryland |
| oak | Oak | hardwood | forest, temperate, coast, river | 0.9-5.0 | 1.15 | 1.18 | 2 | 5 | 5 | 0.20 | 5 | 0.60 | bast_strip | tannin, cordage | hardwood |
| palm | Palm | softwood | island, coast, tropical, delta | 0.6-2.6 | 0.72 | 0.68 | 3 | 2 | 2 | 0.08 | 2 | 0.90 | palm_frond | thatch, weaving | tropical |
| paperbark | Paperbark | softwood | swamp, wetlands, coast, tropical | 0.6-3.0 | 0.78 | 0.72 | 4 | 2 | 2 | 0.25 | 3 | 0.90 | birch_bark | sheet bark, containers | wetland, tropical |
| pine | Pine | resinous | mountain, boreal, forest, dry | 0.8-4.1 | 0.90 | 0.82 | 4 | 2 | 3 | 0.90 | 3 | 1.20 | inner_bark_fiber | tinder, fiber | conifer |
| poplar | Poplar | softwood | forest, river, lake, wetlands | 0.7-3.6 | 0.80 | 0.75 | 3 | 2 | 2 | 0.10 | 2 | 0.80 | inner_bark_fiber | fiber, kindling | hardwood_light |
| redwood | Redwood | softwood | coast, temperate_rainforest, forest | 0.9-5.2 | 0.96 | 0.90 | 3 | 3 | 5 | 0.30 | 5 | 0.85 | cedar_bark | insulation, roofing | conifer, coastal |
| saguaro_rib | Saguaro Rib | driftwood | desert, dry, badlands | 0.3-1.8 | 0.64 | 0.60 | 3 | 2 | 1 | 0.02 | 1 | 1.00 |  | framework, splints | desert |
| spruce | Spruce | resinous | boreal, subarctic, forest, mountain | 0.9-4.5 | 0.92 | 0.85 | 4 | 3 | 4 | 0.85 | 3 | 1.10 | spruce_root | sewing, lashing | conifer |
| sycamore | Sycamore | hardwood | river, wetlands, forest, coast | 0.8-4.4 | 1.03 | 0.98 | 2 | 3 | 4 | 0.08 | 3 | 0.70 | bast_strip | weaving, lining | wetland |
| tamarack | Tamarack | resinous | boreal, wetlands, subarctic, mountain | 0.8-4.2 | 1.00 | 0.94 | 4 | 3 | 4 | 0.72 | 5 | 0.95 | inner_bark_fiber | lashing, pitch | conifer, wetland |
| tamarisk | Tamarisk | hardwood | desert, dry, delta, badlands | 0.6-3.2 | 0.98 | 0.94 | 3 | 3 | 3 | 0.18 | 3 | 0.90 | bast_strip | fencing, twigs | dryland, wetland_edge |
| teak | Teak | hardwood | jungle, tropical, wetlands, coast | 0.9-5.2 | 1.10 | 1.02 | 2 | 5 | 5 | 0.15 | 5 | 0.70 | bast_strip | planking, rafts | tropical, hardwood |
| walnut | Walnut | hardwood | forest, river, temperate, mountain | 0.8-4.8 | 1.14 | 1.11 | 2 | 5 | 4 | 0.07 | 4 | 0.70 | bast_strip | bow staves, tools | hardwood |
| olive | Wild Olive | hardwood | coast, dry, badlands, savanna | 0.7-3.6 | 1.16 | 1.10 | 2 | 5 | 4 | 0.10 | 5 | 0.68 | bast_strip | tool handles, stakes | dryland |
| willow | Willow | softwood | wetlands, swamp, delta, river, lake | 0.6-3.0 | 0.74 | 0.70 | 2 | 2 | 2 | 0.10 | 2 | 0.90 | willow_bark | medicine, withies | wetland |
