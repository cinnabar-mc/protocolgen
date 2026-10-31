// Copyright (c) 2024, The Endstone Project. (https://endstone.dev) All Rights Reserved.
//
// This small CI adapter is derived from Endstone DevTools' exportAll routine.
// It is copied into a source-pinned Endstone checkout and compiled there.
// Keep the Apache-2.0 notice and the upstream attribution with this adapter.

#include "endstone/core/devtools/headless_vanilla_data.h"

#include <filesystem>
#include <fstream>
#include <zstr.hpp>

#include "bedrock/nbt/nbt_io.h"
#include "bedrock/util/string_byte_output.h"
#include "endstone/core/devtools/vanilla_data.h"

namespace fs = std::filesystem;

namespace endstone::core::devtools {

bool exportHeadlessVanillaData(const fs::path &base_path)
{
    auto *data = VanillaData::get();
    if (!data) {
        return false;
    }
    fs::create_directories(base_path);

    const auto save_json = [&](const nlohmann::json &value, const char *name) {
        std::ofstream file(base_path / name);
        file << value;
        file.close();
        return file.good();
    };
    const auto save_nbt = [&](const CompoundTag &value, const char *name) {
        std::string buffer;
        BigEndianStringByteOutput output(buffer);
        NbtIo::writeNamedTag("", value, output);
        std::ofstream file(base_path / name, std::ios::out | std::ios::binary);
        try {
            zstr::ostream compressed(file);
            compressed << buffer;
            compressed.flush();
        }
        catch (const std::ios_base::failure &) {
            return false;
        }
        file.close();
        return file.good();
    };

    if (!save_json(data->block_types, "block_types.json") || !save_json(data->block_states, "block_states.json") ||
        !save_json(data->block_tags, "block_tags.json") || !save_json(data->items, "items.json") ||
        !save_json(data->item_tags, "item_tags.json") || !save_json(data->creative_groups, "creative_groups.json") ||
        !save_json(data->biomes, "biomes.json") || !save_nbt(data->item_components, "item_components.nbt")) {
        return false;
    }

    auto block_palette = CompoundTag();
    block_palette.put("blocks", data->block_palette.copy());
    if (!save_nbt(block_palette, "block_palette.nbt")) {
        return false;
    }

    auto creative_items = CompoundTag();
    creative_items.put("items", data->creative_items.copy());
    if (!save_nbt(creative_items, "creative_items.nbt")) {
        return false;
    }

    const auto recipe_json = nlohmann::json{
        {"shapeless", data->recipes.shapeless},
        {"shaped", data->recipes.shaped},
        {"furnace", data->recipes.furnace},
        {"furnaceAux", data->recipes.furnace_aux},
        {"multi", data->recipes.multi},
        {"userDataShapeless", data->recipes.user_data_shapeless},
        {"shapelessChemistry", data->recipes.shapeless_chemistry},
        {"shapedChemistry", data->recipes.shaped_chemistry},
        {"smithingTransform", data->recipes.smithing_transform},
        {"smithingTrim", data->recipes.smithing_trim},
        {"potionMixes", data->recipes.potion_mixes},
        {"containerMixes", data->recipes.container_mixes},
        {"materialReducer", data->recipes.material_reducer},
    };
    return save_json(recipe_json, "recipes.json");
}

}  // namespace endstone::core::devtools
