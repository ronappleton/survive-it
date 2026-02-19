# Animals

Source: `internal/game/animals.go` (`AnimalCatalog`).

Total animals: **95**.

| ID | Name | Domain | Biome Tags | Weight (kg) | Edible Yield | Nutrition /100g | Disease Risks |
| --- | --- | --- | --- | --- | --- | --- | --- |
| alligator | Alligator | land | swamp, wetlands, delta, coast | 35-450 | 44% | 143kcal 29gP 3gF 0gS | Salmonella risk (10%, any) |
| antelope | Antelope | land | savanna, badlands, dry | 22-140 | 47% | 152kcal 29gP 3gF 0gS | Meat contamination (5%, any) |
| beaver | Beaver | land | lake, delta, river, boreal, forest | 8-32 | 46% | 146kcal 24gP 5gF 0gS | Waterborne contamination (8%, any) |
| bison | Bison | land | savanna, badlands, dry, mountain | 320-950 | 50% | 143kcal 28gP 2gF 0gS | Field contamination (4%, any) |
| black_bear | Black Bear | land | forest, boreal, mountain, coast, lake | 60-300 | 45% | 155kcal 27gP 5gF 0gS | Trichinella risk (14%, muscle) |
| boa_constrictor | Boa Constrictor | land | jungle, wetlands, swamp, tropical, river | 4-35 | 39% | 96kcal 20gP 2gF 0gS | Reptile salmonella (11%, any) |
| brown_bear | Brown Bear | land | boreal, subarctic, arctic, mountain, delta | 110-650 | 45% | 158kcal 27gP 6gF 0gS | Trichinella risk (15%, muscle) |
| caiman | Caiman | land | jungle, wetlands, swamp, delta | 8-250 | 43% | 140kcal 28gP 3gF 0gS | Salmonella risk (11%, any) |
| camel | Camel | land | desert, dry, savanna, badlands | 320-700 | 47% | 148kcal 27gP 4gF 0gS | GI contamination (5%, any) |
| capybara | Capybara | land | wetlands, swamp, jungle, delta | 20-75 | 48% | 172kcal 24gP 8gF 0gS | Waterborne contamination (9%, any) |
| caribou | Caribou | land | arctic, tundra, subarctic, boreal | 60-320 | 47% | 143kcal 29gP 2gF 0gS | Caribou parasites (5%, muscle) |
| cobra | Cobra | land | jungle, savanna, badlands, tropical, wetlands | 0.7-6.8 | 35% | 95kcal 20gP 1gF 0gS | Neurotoxic venom risk (22%, any) |
| cougar | Cougar | land | mountain, forest, coast, badlands | 28-100 | 40% | 164kcal 28gP 5gF 0gS | Predator parasite load (11%, muscle) |
| coyote | Coyote | land | desert, dry, badlands, savanna, mountain | 8-25 | 39% | 160kcal 27gP 5gF 0gS | Scavenger pathogen load (11%, any) |
| crocodile | Crocodile | land | jungle, wetlands, swamp, delta, savanna | 90-900 | 44% | 146kcal 29gP 3gF 0gS | Salmonella risk (12%, any) |
| deer | Deer | land | forest, boreal, coast, lake, mountain | 35-180 | 46% | 158kcal 30gP 3gF 0gS | GI contamination (6%, any) |
| dingo | Dingo | land | savanna, badlands, dry, desert, island | 11-24 | 39% | 158kcal 26gP 5gF 0gS | Pathogen load (10%, any) |
| elk | Elk | land | forest, mountain, boreal, coast | 110-480 | 48% | 150kcal 30gP 2gF 0gS | Field contamination (4%, any) |
| fox | Fox | land | forest, boreal, desert, dry, mountain | 2.5-14 | 38% | 156kcal 26gP 5gF 0gS | Parasite load (10%, muscle) |
| hyena | Hyena | land | savanna, badlands, dry | 35-85 | 43% | 148kcal 27gP 4gF 0gS | Scavenger pathogen load (12%, any) |
| iguana | Iguana | land | jungle, tropical, island, dry_forest | 1.2-6 | 42% | 119kcal 20gP 3gF 0gS | Salmonella risk (11%, any) |
| jackal | Jackal | land | savanna, badlands, dry, desert | 6-16 | 38% | 157kcal 26gP 5gF 0gS | Scavenger pathogen load (11%, any) |
| jaguar | Jaguar | land | jungle, wetlands, swamp, delta | 45-120 | 41% | 167kcal 28gP 6gF 0gS | Predator parasite load (11%, muscle) |
| kangaroo | Kangaroo | land | savanna, badlands, dry, desert | 18-90 | 46% | 98kcal 22gP 2gF 0gS | Field contamination (5%, any) |
| leopard | Leopard | land | savanna, jungle, badlands, dry, mountain | 25-90 | 41% | 167kcal 28gP 6gF 0gS | Predator parasite load (11%, muscle) |
| lion | Lion | land | savanna, badlands, dry | 110-250 | 42% | 170kcal 28gP 6gF 0gS | Predator parasite load (12%, muscle) |
| monitor_lizard | Monitor Lizard | land | savanna, jungle, wetlands, swamp, island | 3-65 | 41% | 126kcal 22gP 4gF 0gS | Reptile salmonella (13%, any) |
| moose | Moose | land | boreal, subarctic, lake, delta | 180-550 | 49% | 146kcal 29gP 2gF 0gS | Meat contamination (3%, any) |
| mountain_goat | Mountain Goat | land | mountain, highland, montane | 50-120 | 46% | 142kcal 27gP 3gF 0gS | GI contamination (5%, any) |
| mouse | Mouse | land | forest, boreal, jungle, desert, savanna, wetlands, swamp, mountain, coast, arctic | 0.015-0.045 | 40% | 120kcal 20gP 5gF 0gS | Hantavirus risk (5%, respiratory); Salmonella risk (8%, any) |
| muskrat | Muskrat | land | wetlands, swamp, delta, lake, river | 0.6-2.3 | 44% | 138kcal 21gP 5gF 0gS | Leptospirosis risk (9%, blood) |
| python | Python | land | jungle, swamp, wetlands, tropical, island | 8-95 | 40% | 96kcal 20gP 2gF 0gS | Reptile salmonella (12%, any) |
| rabbit | Rabbit | land | forest, mountain, coast, savanna, badlands, dry | 1-2.8 | 52% | 173kcal 33gP 4gF 0gS | Liver worms (8%, liver); Tularemia risk (4%, blood) |
| rattlesnake | Rattlesnake | land | desert, dry, savanna, badlands | 0.3-2.8 | 36% | 93kcal 20gP 1gF 0gS | Venom contamination (18%, any) |
| scorpion | Scorpion | land | desert, dry, savanna, badlands, tropical_dry_forest | 0.01-0.09 | 30% | 121kcal 22gP 3gF 0gS | Venom contamination (15%, any) |
| sea_snake | Sea Snake | land | coast, island, delta, tropical | 0.4-2.5 | 34% | 94kcal 19gP 2gF 0gS | Venom contamination (24%, any) |
| tarantula | Tarantula | land | jungle, savanna, badlands, desert, dry | 0.02-0.18 | 32% | 135kcal 25gP 4gF 0gS | Venom contamination (12%, any) |
| tiger | Tiger | land | jungle, swamp, wetlands, tropical | 80-310 | 42% | 168kcal 28gP 6gF 0gS | Predator parasite load (12%, muscle) |
| warthog | Warthog | land | savanna, badlands, dry | 35-150 | 45% | 163kcal 26gP 6gF 0gS | Trichinella (8%, muscle) |
| boar | Wild Boar | land | forest, jungle, wetlands, swamp, island | 25-160 | 47% | 160kcal 27gP 5gF 0gS | Trichinella (7%, muscle) |
| wolf | Wolf | land | boreal, subarctic, forest, mountain, tundra | 25-75 | 41% | 162kcal 27gP 5gF 0gS | Predator parasite load (10%, muscle) |
| wolf_spider | Wolf Spider | land | forest, jungle, savanna, badlands, swamp | 0.005-0.03 | 28% | 128kcal 24gP 3gF 0gS | Venom contamination (10%, any) |
| anchovy | Anchovy | water | coast, delta, island | 0.005-0.08 | 58% | 131kcal 20gP 4gF 0gS | Marine contamination (4%, any) |
| arapaima | Arapaima | water | jungle, wetlands, swamp, delta, river | 8-200 | 61% | 121kcal 22gP 3gF 0gS | Fish parasite (6%, muscle) |
| char | Arctic Char | water | arctic, tundra, subarctic, lake, delta | 0.3-8 | 57% | 172kcal 21gP 9gF 0gS | Cold-water parasite (4%, muscle) |
| barracuda | Barracuda | water | coast, island, tropical | 1.5-45 | 57% | 124kcal 21gP 3gF 0gS | Ciguatera risk (8%, muscle) |
| bluegill | Bluegill | water | lake, river, wetlands, swamp, delta | 0.08-0.9 | 50% | 120kcal 21gP 3gF 0gS | Fish parasite (6%, muscle) |
| carp | Carp | water | lake, delta, river, wetlands, savanna | 0.9-35 | 54% | 127kcal 18gP 5gF 0gS | Water contamination (8%, any) |
| catfish | Catfish | water | wetlands, swamp, jungle, delta, savanna | 0.4-9 | 52% | 144kcal 18gP 8gF 0gS | Waterborne contamination (8%, any) |
| catla | Catla | water | river, delta, wetlands, jungle, tropical | 1-45 | 55% | 127kcal 19gP 5gF 0gS | Water contamination (8%, any) |
| cod | Cod | water | coast, delta, island | 0.8-40 | 59% | 105kcal 23gP 1gF 0gS | Marine parasite (5%, muscle) |
| crappie | Crappie | water | lake, river, wetlands, delta | 0.1-1.5 | 52% | 118kcal 20gP 3gF 0gS | Fish parasite (6%, muscle) |
| eel | Eel | water | coast, delta, swamp, river, lake | 0.2-8 | 52% | 184kcal 18gP 12gF 0gS | Fish parasite (5%, muscle) |
| flounder | Flounder | water | coast, delta, island | 0.2-8 | 56% | 117kcal 24gP 2gF 0gS | Marine parasite (5%, muscle) |
| grayling | Grayling | water | river, subarctic, boreal, mountain, lake | 0.15-2.6 | 54% | 122kcal 20gP 4gF 0gS | Fish parasite (5%, muscle) |
| grouper_fish | Grouper | water | coast, island, tropical | 1.2-110 | 59% | 118kcal 24gP 2gF 0gS | Ciguatera risk (5%, muscle) |
| halibut | Halibut | water | coast, island, delta | 2-220 | 63% | 111kcal 23gP 2gF 0gS | Marine parasite (5%, muscle) |
| herring | Herring | water | coast, delta, island | 0.04-0.7 | 57% | 158kcal 18gP 9gF 0gS | Marine parasite (5%, muscle) |
| largemouth_bass | Largemouth Bass | water | lake, river, wetlands, swamp, delta | 0.4-10 | 55% | 130kcal 20gP 5gF 0gS | Fish parasite (6%, muscle) |
| mackerel | Mackerel | water | coast, island, delta | 0.15-4 | 58% | 205kcal 19gP 13gF 0gS | Scombroid contamination (6%, any) |
| mahi_mahi | Mahi-Mahi | water | coast, island, tropical | 1.5-38 | 61% | 134kcal 23gP 4gF 0gS | Histamine contamination (5%, any) |
| northern_pike | Northern Pike | water | lake, river, boreal, subarctic, delta | 0.8-21 | 55% | 113kcal 20gP 2gF 0gS | Freshwater parasite (6%, muscle) |
| perch | Perch | water | lake, river, delta, forest, boreal | 0.15-2.2 | 52% | 117kcal 24gP 2gF 0gS | Fish parasite (5%, muscle) |
| piranha | Piranha | water | jungle, wetlands, swamp, delta, river | 0.2-2 | 47% | 119kcal 21gP 3gF 0gS | Fish parasite (7%, muscle) |
| red_snapper | Red Snapper | water | coast, island, delta, tropical | 0.6-18 | 56% | 128kcal 26gP 2gF 0gS | Ciguatera risk (5%, muscle) |
| rohu | Rohu | water | river, delta, wetlands, jungle, tropical | 0.8-18 | 55% | 123kcal 20gP 4gF 0gS | Water contamination (8%, any) |
| salmon | Salmon | water | coast, river, delta, temperate_rainforest, island | 1.2-18 | 58% | 206kcal 22gP 12gF 0gS | Marine parasite (4%, muscle) |
| sardine | Sardine | water | coast, island, delta | 0.02-0.18 | 61% | 208kcal 25gP 11gF 0gS | Marine contamination (5%, any) |
| sturgeon | Sturgeon | water | delta, river, coast, lake | 3-250 | 60% | 135kcal 19gP 6gF 0gS | Fish parasite (5%, muscle) |
| swordfish | Swordfish | water | coast, island | 20-540 | 62% | 172kcal 20gP 9gF 0gS | Marine parasite (5%, muscle) |
| tarpon | Tarpon | water | coast, delta, island, wetlands | 3-130 | 56% | 126kcal 22gP 3gF 0gS | Marine parasite (5%, muscle) |
| tilapia | Tilapia | water | jungle, wetlands, swamp, savanna, delta | 0.2-4.5 | 54% | 129kcal 20gP 3gF 0gS | Water contamination (8%, any) |
| trout | Trout | water | lake, river, mountain, forest, boreal | 0.25-4.5 | 56% | 141kcal 20gP 6gF 0gS | Fish parasite (5%, muscle) |
| tuna | Tuna | water | coast, island | 4-250 | 63% | 132kcal 29gP 1gF 0gS | Histamine contamination (4%, any) |
| walleye | Walleye | water | lake, river, delta, boreal, subarctic | 0.4-9 | 56% | 111kcal 21gP 2gF 0gS | Freshwater parasite (5%, muscle) |
| albatross | Albatross | air | coast, island | 2.5-12 | 47% | 178kcal 23gP 8gF 0gS | Bird-borne contamination (8%, any) |
| cormorant | Cormorant | air | coast, island, delta, lake | 1.2-4 | 45% | 174kcal 24gP 8gF 0gS | Bird-borne contamination (9%, any) |
| crow | Crow | air | forest, savanna, badlands, coast, wetlands | 0.25-0.65 | 40% | 186kcal 23gP 9gF 0gS | Scavenger contamination (12%, any) |
| dove | Dove | air | forest, savanna, badlands, dry, coast | 0.09-0.35 | 40% | 172kcal 23gP 8gF 0gS | Bird-borne contamination (7%, any) |
| duck | Duck | air | lake, delta, wetlands, swamp, coast | 0.7-2.1 | 49% | 337kcal 19gP 28gF 0gS | Salmonella risk (8%, any) |
| egret | Egret | air | wetlands, swamp, delta, coast, jungle | 0.7-1.8 | 45% | 166kcal 23gP 7gF 0gS | Bird-borne contamination (8%, any) |
| goose | Goose | air | lake, delta, wetlands, coast, tundra | 2-7 | 51% | 305kcal 25gP 21gF 0gS | Salmonella risk (9%, any) |
| grouse | Grouse | air | boreal, subarctic, forest, mountain | 0.55-1.4 | 50% | 190kcal 25gP 9gF 0gS | GI contamination (5%, any) |
| heron | Heron | air | wetlands, swamp, delta, lake, coast | 1-3.8 | 46% | 170kcal 24gP 7gF 0gS | Bird-borne contamination (8%, any) |
| kingfisher | Kingfisher | air | river, lake, delta, wetlands, jungle | 0.03-0.25 | 38% | 171kcal 23gP 8gF 0gS | Bird-borne contamination (8%, any) |
| macaw | Macaw | air | jungle, tropical, island, rainforest | 0.8-1.8 | 43% | 205kcal 23gP 11gF 0gS | Bird-borne GI contamination (7%, any) |
| partridge | Partridge | air | forest, mountain, savanna, badlands, dry | 0.25-0.9 | 46% | 180kcal 24gP 8gF 0gS | Bird-borne contamination (7%, any) |
| pelican | Pelican | air | coast, island, delta, wetlands | 2.5-14 | 46% | 176kcal 23gP 8gF 0gS | Bird-borne contamination (9%, any) |
| pheasant | Pheasant | air | forest, savanna, badlands, mountain, coast | 0.7-2 | 48% | 181kcal 25gP 8gF 0gS | GI contamination (6%, any) |
| ptarmigan | Ptarmigan | air | arctic, tundra, subarctic, boreal | 0.35-0.95 | 47% | 183kcal 24gP 8gF 0gS | GI contamination (5%, any) |
| quail | Quail | air | savanna, badlands, dry, forest, coast | 0.12-0.35 | 44% | 134kcal 21gP 5gF 0gS | Salmonella risk (8%, any) |
| raven | Raven | air | boreal, subarctic, mountain, coast, tundra | 0.4-1.6 | 42% | 188kcal 23gP 9gF 0gS | Scavenger contamination (12%, any) |
| pigeon | Rock Pigeon | air | desert, dry, savanna, badlands, coast | 0.25-0.42 | 45% | 213kcal 24gP 13gF 0gS | Bird-borne GI contamination (9%, any) |
| seagull | Seagull | air | coast, island, delta, lake | 0.35-1.8 | 42% | 202kcal 22gP 11gF 0gS | Scavenger contamination (11%, any) |
| wild_turkey | Wild Turkey | air | forest, mountain, wetlands, savanna | 2.5-11 | 52% | 189kcal 29gP 7gF 0gS | Salmonella risk (9%, any) |
