<script lang="ts">
    import {
        Action,
        Biome,
        DraftKind,
        GamePhase,
        Structure,
        type DraftItem,
        type GameState,
    } from "../lib/types";
    import { debugLog } from "../lib/debug";

    export let game: GameState;
    export let playerId = 0;
    export let role: "player" | "spectator" | "" = "";

    export let onPick: (marketIndex: number) => void;

    $: canPick =
        role === "player" &&
        game.CurrentPhase === GamePhase.Pick &&
        game.CurrentPlayer === playerId;

    $: debugLog("market.state", {
        role,
        playerId,
        currentPlayer: game.CurrentPlayer,
        currentPhase: game.CurrentPhase,
        canPick,
        marketCount: game.Market.length,
    });

    function biomeName(biome: Biome) {
        switch (biome) {
            case Biome.Forest:
                return "Forest";
            case Biome.Mountain:
                return "Mountain";
            case Biome.Plain:
                return "Plain";
            case Biome.River:
                return "River";
            default:
                return "Unknown";
        }
    }

    function structureName(structure: Structure) {
        switch (structure) {
            case Structure.Bridge:
                return "Bridge";
            case Structure.Watchtower:
                return "Watchtower";
            case Structure.Road:
                return "Road";
            case Structure.Outpost:
                return "Outpost";
            case Structure.City:
                return "City";
            default:
                return "Structure";
        }
    }

    function actionName(action: Action) {
        switch (action) {
            case Action.Harvest:
                return "Harvest";
            case Action.Reinforce:
                return "Reinforce";
            case Action.Expansion:
                return "Expansion";
            default:
                return "Action";
        }
    }

    function itemTitle(item: DraftItem) {
        switch (item.Kind) {
            case DraftKind.Tile:
                return biomeName(item.Biome);
            case DraftKind.Upgrade:
                return "Upgrade";
            case DraftKind.Structure:
                return structureName(item.Structure);
            case DraftKind.Action:
                return actionName(item.Action);
            default:
                return "Unknown";
        }
    }

    function itemType(item: DraftItem) {
        switch (item.Kind) {
            case DraftKind.Tile:
                return "Tile";
            case DraftKind.Upgrade:
                return "Upgrade";
            case DraftKind.Structure:
                return "Structure";
            case DraftKind.Action:
                return "Action";
            default:
                return "";
        }
    }

    function itemIcon(item: DraftItem) {
        if (item.Kind === DraftKind.Tile) {
            switch (item.Biome) {
                case Biome.Forest:
                    return "♣";
                case Biome.Mountain:
                    return "▲";
                case Biome.Plain:
                    return "◆";
                case Biome.River:
                    return "≈";
            }
        }

        if (item.Kind === DraftKind.Upgrade) {
            return "↑";
        }

        if (item.Kind === DraftKind.Structure) {
            switch (item.Structure) {
                case Structure.Outpost:
                    return "⌂";
                case Structure.City:
                    return "▦";
                case Structure.Road:
                    return "━";
                case Structure.Bridge:
                    return "⌒";
                case Structure.Watchtower:
                    return "♜";
                default:
                    return "■";
            }
        }

        if (item.Kind === DraftKind.Action) {
            switch (item.Action) {
                case Action.Harvest:
                    return "✦";
                case Action.Reinforce:
                    return "+";
                case Action.Expansion:
                    return "⇱";
                default:
                    return "!";
            }
        }

        return "?";
    }

    function itemDescription(item: DraftItem) {
        switch (item.Kind) {
            case DraftKind.Tile:
                switch (item.Biome) {
                    case Biome.Forest:
                        return "Place a forest tile. Forest outposts project influence farther.";
                    case Biome.Mountain:
                        return "Place a mountain tile. Connected mountains improve stone production.";
                    case Biome.Plain:
                        return "Place a plain tile. Cities can only be built on plains.";
                    case Biome.River:
                        return "Place a river tile. Rivers block normal structures and create bridge points.";
                    default:
                        return "Place this hex adjacent to the existing map.";
                }

            case DraftKind.Upgrade:
                return "Upgrade one controlled non-river tile. Upgraded tiles produce more resources.";

            case DraftKind.Structure:
                if (item.Structure === Structure.Bridge) {
                    return "Build on a river adjacent to your territory. Helps control crossings.";
                }

                if (item.Structure === Structure.Watchtower) {
                    return "Build on controlled land. Projects influence up to distance 2.";
                }

                if (item.Structure === Structure.Road) {
                    return "Build on controlled land. Connects infrastructure and helps with road scoring.";
                }

                if (item.Structure === Structure.Outpost) {
                    return "Build on valid land to claim and project influence.";
                }

                if (item.Structure === Structure.City) {
                    return "Cities score points and project strong influence. Usually created from outposts.";
                }

                return "Build this structure on a valid tile.";

            case DraftKind.Action:
                if (item.Action === Action.Harvest) {
                    return "Choose one controlled non-river tile and gain extra resources from it.";
                }

                if (item.Action === Action.Reinforce) {
                    return "Add temporary influence to one tile for the next recount.";
                }

                if (item.Action === Action.Expansion) {
                    return "Gain 1 Wood and 1 Grain immediately.";
                }

                return "Use this special action.";

            default:
                return "";
        }
    }

    function itemClass(item: DraftItem) {
        if (item.Kind === DraftKind.Tile) {
            switch (item.Biome) {
                case Biome.Forest:
                    return "bg-[#5b9368] border-[#2f6546]";
                case Biome.Mountain:
                    return "bg-[#a8adb2] border-[#656b73]";
                case Biome.Plain:
                    return "bg-[#d9b56a] border-[#9b7034]";
                case Biome.River:
                    return "bg-[#6eb8c5] border-[#327b8d]";
            }
        }

        if (item.Kind === DraftKind.Upgrade) {
            return "bg-[#f2c36b] border-[#9b7034]";
        }

        if (item.Kind === DraftKind.Structure) {
            return "bg-[#ead7aa] border-[#6b4a2f]";
        }

        if (item.Kind === DraftKind.Action) {
            return "bg-[#73c4bd] border-[#327b8d]";
        }

        return "bg-[#ead7aa] border-[#6b4a2f]";
    }
</script>

<section
    class="rounded-3xl bg-[#23444c] p-5 shadow-md ring-1 ring-[#f8efe0]/10"
>
    <div class="flex items-center justify-between gap-3">
        <h2 class="text-xl font-black text-[#fff7e8]">Market</h2>

        <div
            class={[
                "rounded-xl px-3 py-1 text-xs font-black uppercase tracking-wider",
                canPick
                    ? "bg-[#f2c36b] text-[#142833]"
                    : "bg-[#f8efe0]/10 text-[#9fc9c5]",
            ].join(" ")}
        >
            {canPick ? "Pick 1" : `${game.Market.length} cards`}
        </div>
    </div>

    <div class="mt-4 grid grid-cols-2 gap-3">
        {#each game.Market as item, index}
            <button
                class={[
                    "group relative h-32 cursor-pointer rounded-2xl border-2 p-3 text-left text-[#142833] shadow-[0_6px_0_rgba(0,0,0,0.18)] transition",
                    "hover:-translate-y-1 hover:brightness-105 active:translate-y-1",
                    "disabled:cursor-not-allowed disabled:opacity-55 disabled:hover:translate-y-0 disabled:hover:brightness-100",
                    itemClass(item),
                ].join(" ")}
                type="button"
                disabled={!canPick}
                aria-label={`${itemTitle(item)} ${itemType(item)}. ${itemDescription(item)}`}
                on:click={() => {
                    debugLog("market.pick.click", {
                        index,
                        canPick,
                        role,
                        playerId,
                        currentPlayer: game.CurrentPlayer,
                        currentPhase: game.CurrentPhase,
                        item,
                    });

                    onPick(index);
                }}
            >
                <div class="flex h-full flex-col justify-between">
                    <div class="flex items-start justify-between gap-2">
                        <div
                            class="grid h-10 w-10 place-items-center rounded-xl bg-white/35 text-2xl font-black"
                        >
                            {itemIcon(item)}
                        </div>

                        <div
                            class="rounded-lg bg-white/30 px-2 py-1 text-[10px] font-black uppercase tracking-wider"
                        >
                            {itemType(item)}
                        </div>
                    </div>

                    <div>
                        <div class="text-lg font-black leading-5">
                            {itemTitle(item)}
                        </div>
                    </div>
                </div>

                <div
                    class="pointer-events-none absolute bottom-[calc(100%+10px)] left-1/2 z-50 hidden w-64 -translate-x-1/2 rounded-2xl bg-[#142833] p-4 text-[#f8efe0] shadow-xl ring-1 ring-white/10 group-hover:block"
                >
                    <div class="flex items-center justify-between gap-3">
                        <div class="text-sm font-black">
                            {itemTitle(item)}
                        </div>

                        <div
                            class="rounded-lg bg-[#f8efe0]/10 px-2 py-1 text-[10px] font-black uppercase tracking-wider text-[#9fc9c5]"
                        >
                            {itemType(item)}
                        </div>
                    </div>

                    <div
                        class="mt-2 text-xs font-semibold leading-5 text-[#d9e6df]"
                    >
                        {itemDescription(item)}
                    </div>

                    {#if canPick}
                        <div
                            class="mt-3 text-[10px] font-black uppercase tracking-[0.18em] text-[#f2c36b]"
                        >
                            Click to draft
                        </div>
                    {/if}
                </div>
            </button>
        {/each}
    </div>
</section>
