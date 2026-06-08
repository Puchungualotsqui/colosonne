<script lang="ts">
    import ResourceIcon from "./ResourceIcon.svelte";
    import CostBadge from "./CostBadge.svelte";
    import Board from "./Board.svelte";
    import Market from "./Market.svelte";
    import HandCard from "./HandCard.svelte";
    import MiniHandCard from "./MiniHandCard.svelte";
    import CardTooltip from "./CardTooltip.svelte";
    import BuildActionTooltip from "./BuildActionTooltip.svelte";
    import {
        Action,
        Biome,
        DraftKind,
        GamePhase,
        Structure,
        type BuildAction,
        type BuildCostsByPlayer,
        type DraftItem,
        type GameState,
        type Player,
        type ResourceCostResponse,
        type TargetBuildAction,
    } from "../lib/types";
    import { debugLog } from "../lib/debug";

    export let game: GameState;
    export let roomId = "";
    export let playerId = 0;
    export let role: "player" | "spectator" | "" = "";
    export let error = "";
    export let buildCosts: BuildCostsByPlayer = {};

    export let onPick: (marketIndex: number) => void;
    export let onPlaceTile: (handIndex: number, x: number, y: number) => void;
    export let onUseDraft: (
        handIndex: number,
        x: number,
        y: number,
        targetPlayerId?: number,
    ) => void;
    export let onDiscardDraft: (handIndex: number) => void;
    export let onPassPlace: () => void;
    export let onBuild: (action: BuildAction, x: number, y: number) => void;
    export let onPassBuild: () => void;
    export let onLeaveRoom: () => void;
    export let onCopyRoomCode: () => void;

    type Cost = {
        wood?: number;
        stone?: number;
        grain?: number;
        relic?: number;
    };

    const RES_WOOD = 1;
    const RES_STONE = 2;
    const RES_GRAIN = 3;
    const RES_RELIC = 4;

    let selectedBuildAction: TargetBuildAction | null = null;
    let selectedHandIndex = -1;

    let previousResourcesByPlayer = new Map<number, string>();
    let gainedResourcesByPlayer = new Map<number, Map<number, number>>();

    $: me = game.Players.find((p) => p.Id === playerId);
    $: myHand = me?.Hand ?? [];
    $: selectedHandItem =
        selectedHandIndex >= 0 ? (myHand[selectedHandIndex] ?? null) : null;

    $: isMyTurn = role === "player" && game.CurrentPlayer === playerId;
    $: currentPhaseName = phaseName(game.CurrentPhase);
    $: currentPlayerName =
        game.CurrentPlayer === playerId
            ? "You"
            : `Player ${game.CurrentPlayer}`;

    $: if (game.CurrentPhase !== GamePhase.Build) {
        selectedBuildAction = null;
    }

    $: {
        if (
            !isMyTurn ||
            game.CurrentPhase !== GamePhase.Place ||
            myHand.length === 0
        ) {
            selectedHandIndex = -1;
        } else if (
            selectedHandIndex < 0 ||
            selectedHandIndex >= myHand.length
        ) {
            selectedHandIndex = firstUsableHandIndex();
        }
    }

    $: selectedHandIsUsable = canUseHandLocally(selectedHandItem);
    $: canDiscardSelected =
        isMyTurn &&
        game.CurrentPhase === GamePhase.Place &&
        selectedHandIndex >= 0 &&
        selectedHandItem !== null &&
        !selectedHandIsUsable;

    $: trackResourceGains(game);

    let outpostCost: Cost = {};
    let cityCost: Cost = {};
    let settlementCost: Cost = {};
    let blockadeCost: Cost = {};
    let floodworksCost: Cost = {};

    $: myBuildCosts = buildCosts[String(playerId)];

    $: outpostCost = costFromServer(myBuildCosts?.outpost, {
        wood: 2 + activeBuiltCount(Structure.Outpost),
        stone: 1,
    });

    $: cityCost = costFromServer(myBuildCosts?.city, {
        stone: 2,
        grain: 3 + activeBuiltCount(Structure.City),
    });

    $: settlementCost = costFromServer(myBuildCosts?.settlement, {
        wood: 2,
        stone: 2,
        grain: 2 + activeBuiltCount(Structure.Settlement),
    });

    $: blockadeCost = costFromServer(myBuildCosts?.blockade, {
        wood: 1,
        grain: 1,
    });

    $: floodworksCost = costFromServer(myBuildCosts?.floodworks, {
        relic: 3 + (me?.FloodworksBought ?? 0) * 2,
    });

    $: canAffordOutpost = canPay(me, outpostCost);
    $: canAffordCity = canPay(me, cityCost);
    $: canAffordSettlement = canPay(me, settlementCost);
    $: canAffordBlockade = canPay(me, blockadeCost);
    $: canAffordFloodworks = canPay(me, floodworksCost);

    $: hasOutpostTarget = game.Map.some((tile) =>
        isBuildTarget("outpost", tile),
    );
    $: hasCityTarget = game.Map.some((tile) => isBuildTarget("city", tile));
    $: hasSettlementTarget = game.Map.some((tile) =>
        isBuildTarget("settlement", tile),
    );
    $: hasBlockadeTarget = game.Map.some((tile) =>
        isBuildTarget("blockade", tile),
    );
    $: hasFloodTarget = game.Map.some((tile) => isBuildTarget("flood", tile));

    $: canSelectOutpost = canAffordOutpost && hasOutpostTarget;
    $: canSelectCity = canAffordCity && hasCityTarget;
    $: canSelectSettlement = canAffordSettlement && hasSettlementTarget;
    $: canSelectBlockade = canAffordBlockade && hasBlockadeTarget;
    $: canBuyFloodworks = canAffordFloodworks;
    $: canUseFloodToken = (me?.FloodTokens ?? 0) > 0 && hasFloodTarget;

    $: myWoodGain = resourceGainAmount(playerId, RES_WOOD);
    $: myStoneGain = resourceGainAmount(playerId, RES_STONE);
    $: myGrainGain = resourceGainAmount(playerId, RES_GRAIN);
    $: myRelicGain = resourceGainAmount(playerId, RES_RELIC);

    $: hasMyResourceGain =
        myWoodGain > 0 || myStoneGain > 0 || myGrainGain > 0 || myRelicGain > 0;

    function phaseName(phase: GamePhase) {
        switch (phase) {
            case GamePhase.Pick:
                return "Draft";
            case GamePhase.Place:
                return "Use";
            case GamePhase.Build:
                return "Build";
            default:
                return "Unknown";
        }
    }

    function resourceAmount(player: Player | undefined, resourceId: number) {
        return player?.Resources?.[resourceId] ?? 0;
    }

    function resourceKey(player: Player) {
        return JSON.stringify({
            wood: resourceAmount(player, RES_WOOD),
            stone: resourceAmount(player, RES_STONE),
            grain: resourceAmount(player, RES_GRAIN),
            relic: resourceAmount(player, RES_RELIC),
        });
    }

    function trackResourceGains(currentGame: GameState) {
        const nextPrevious = new Map(previousResourcesByPlayer);
        const nextGained = new Map<number, Map<number, number>>();

        for (const player of currentGame.Players) {
            const oldRaw = previousResourcesByPlayer.get(player.Id);
            const newRaw = resourceKey(player);

            if (oldRaw) {
                const oldValue = JSON.parse(oldRaw);
                const gained = new Map<number, number>();

                const woodDelta =
                    resourceAmount(player, RES_WOOD) - oldValue.wood;
                const stoneDelta =
                    resourceAmount(player, RES_STONE) - oldValue.stone;
                const grainDelta =
                    resourceAmount(player, RES_GRAIN) - oldValue.grain;
                const relicDelta =
                    resourceAmount(player, RES_RELIC) - oldValue.relic;

                if (woodDelta > 0) gained.set(RES_WOOD, woodDelta);
                if (stoneDelta > 0) gained.set(RES_STONE, stoneDelta);
                if (grainDelta > 0) gained.set(RES_GRAIN, grainDelta);
                if (relicDelta > 0) gained.set(RES_RELIC, relicDelta);

                if (gained.size > 0) {
                    nextGained.set(player.Id, gained);
                }
            }

            nextPrevious.set(player.Id, newRaw);
        }

        previousResourcesByPlayer = nextPrevious;
        gainedResourcesByPlayer = nextGained;
    }

    function resourcePulse(targetPlayerId: number, resourceId: number) {
        return resourceGainAmount(targetPlayerId, resourceId) > 0;
    }

    function resourceGainAmount(targetPlayerId: number, resourceId: number) {
        return (
            gainedResourcesByPlayer.get(targetPlayerId)?.get(resourceId) ?? 0
        );
    }

    function structureActiveForPlayer(tile: {
        HasOwner: boolean;
        Owner: number;
        StructureOwner: number;
    }) {
        return tile.HasOwner && tile.Owner === tile.StructureOwner;
    }

    function activeBuiltCount(structure: Structure) {
        return game.Map.filter(
            (tile) =>
                tile.Structure === structure &&
                tile.StructureOwner === playerId &&
                structureActiveForPlayer(tile),
        ).length;
    }

    function canPay(player: Player | undefined, cost: Cost) {
        if (!player) return false;

        return (
            resourceAmount(player, RES_WOOD) >= (cost.wood ?? 0) &&
            resourceAmount(player, RES_STONE) >= (cost.stone ?? 0) &&
            resourceAmount(player, RES_GRAIN) >= (cost.grain ?? 0) &&
            resourceAmount(player, RES_RELIC) >= (cost.relic ?? 0)
        );
    }

    function costFromServer(
        response: ResourceCostResponse | undefined,
        fallback: Cost,
    ): Cost {
        if (!response) return fallback;

        return {
            wood: response.wood ?? 0,
            stone: response.stone ?? 0,
            grain: response.grain ?? 0,
            relic: response.relic ?? 0,
        };
    }

    function costTitle(cost: Cost, affordable: boolean, fallback: string) {
        if (affordable) return fallback;

        const parts = [];
        if (cost.wood) parts.push(`${cost.wood} Wood`);
        if (cost.stone) parts.push(`${cost.stone} Stone`);
        if (cost.grain) parts.push(`${cost.grain} Grain`);
        if (cost.relic) parts.push(`${cost.relic} Relic`);

        return `Need ${parts.join(", ")}`;
    }

    function handItems(player: Player | undefined) {
        return player?.Hand ?? [];
    }

    function firstUsableHandIndex() {
        const usableIndex = myHand.findIndex((item) => canUseHandLocally(item));
        if (usableIndex >= 0) return usableIndex;
        return myHand.length > 0 ? 0 : -1;
    }

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
            case Biome.Ruins:
                return "Ruins";
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
            case Structure.Outpost:
                return "Outpost";
            case Structure.City:
                return "City";
            case Structure.Settlement:
                return "Settlement";
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
            case Action.Raid:
                return "Raid";
            default:
                return "Action";
        }
    }

    function draftName(item: DraftItem | null | undefined) {
        if (!item) return "Empty";

        switch (item.Kind) {
            case DraftKind.Tile:
                return `${biomeName(item.Biome)} Tile`;
            case DraftKind.Structure:
                return structureName(item.Structure);
            case DraftKind.Action:
                return actionName(item.Action);
            default:
                return "Unknown";
        }
    }

    function playerColor(targetPlayerId: number) {
        if (targetPlayerId === 1) return "bg-[#1d4e89]";
        if (targetPlayerId === 2) return "bg-[#b94b3f]";
        return "bg-[#6b4a2f]";
    }

    function hexNeighbors(x: number, y: number) {
        return [
            { x: x + 1, y },
            { x: x + 1, y: y - 1 },
            { x, y: y - 1 },
            { x: x - 1, y },
            { x: x - 1, y: y + 1 },
            { x, y: y + 1 },
        ];
    }

    function tileAt(x: number, y: number) {
        return game.Map.find((t) => t.X === x && t.Y === y);
    }

    function controlsTile(
        tile: { HasOwner: boolean; Owner: number } | undefined,
    ) {
        return !!tile && tile.HasOwner && tile.Owner === playerId;
    }

    function hasAdjacentControlledTile(x: number, y: number) {
        return hexNeighbors(x, y).some((n) => controlsTile(tileAt(n.x, n.y)));
    }

    function isEnemyControlledTile(tile: any) {
        return !!tile && tile.HasOwner && tile.Owner !== playerId;
    }

    function isUnownedTile(tile: any) {
        return !!tile && !tile.HasOwner;
    }

    function canBuildBlockadeOnTile(tile: any) {
        if (!tile) return false;

        if (tile.Biome === Biome.River) return false;
        if (tile.HasBlockade) return false;
        if (controlsTile(tile)) return false;

        if (isEnemyControlledTile(tile)) return true;

        if (isUnownedTile(tile)) {
            return hasAdjacentControlledTile(tile.X, tile.Y);
        }

        return false;
    }

    function canUseHandLocally(item: DraftItem | null | undefined) {
        if (!item) return false;

        switch (item.Kind) {
            case DraftKind.Tile:
                return game.Map.some((tile) =>
                    hexNeighbors(tile.X, tile.Y).some((n) => !tileAt(n.x, n.y)),
                );

            case DraftKind.Structure:
                return game.Map.some((tile) =>
                    canUseStructureOnTileLocally(item.Structure, tile),
                );

            case DraftKind.Action:
                if (item.Action === Action.Expansion) return true;
                if (item.Action === Action.Raid)
                    return game.Players.some((p) => p.Id !== playerId);

                if (item.Action === Action.Reinforce)
                    return game.Map.length > 0;

                if (item.Action === Action.Harvest) {
                    return game.Map.some(
                        (tile) =>
                            controlsTile(tile) && tile.Biome !== Biome.River,
                    );
                }

                return false;

            default:
                return false;
        }
    }

    function canUseStructureOnTileLocally(structure: Structure, tile: any) {
        if (!tile) return false;
        if (tile.Structure !== Structure.None) return false;

        switch (structure) {
            case Structure.Bridge:
                return (
                    tile.Biome === Biome.River &&
                    hasAdjacentControlledTile(tile.X, tile.Y)
                );

            case Structure.Watchtower:
                return (
                    tile.Biome !== Biome.River &&
                    tile.Structure === Structure.None &&
                    !isEnemyControlledTile(tile)
                );

            case Structure.Outpost:
            case Structure.City:
            case Structure.Settlement:
                return false;

            default:
                return false;
        }
    }

    function isBuildTarget(action: TargetBuildAction, tile: any) {
        if (!tile) return false;

        switch (action) {
            case "outpost":
                return (
                    tile.Biome !== Biome.River &&
                    tile.Structure === Structure.None
                );

            case "settlement":
                return (
                    tile.Biome !== Biome.River &&
                    tile.Structure === Structure.None &&
                    controlsTile(tile)
                );

            case "blockade":
                return canBuildBlockadeOnTile(tile);

            case "city":
                return (
                    tile.Structure === Structure.Outpost &&
                    tile.StructureOwner === playerId
                );

            case "flood":
                return (
                    tile.Biome !== Biome.River &&
                    tile.Structure === Structure.None
                );

            default:
                return false;
        }
    }

    function handleBuild(action: TargetBuildAction, x: number, y: number) {
        debugLog("gameview.build.received_from_board", {
            action,
            x,
            y,
            playerId,
            currentPlayer: game.CurrentPlayer,
            currentPhase: game.CurrentPhase,
            selectedBuildAction,
            resources: me?.Resources,
            backendCostsForMe: myBuildCosts,
            shownCosts: {
                outpost: outpostCost,
                city: cityCost,
                settlement: settlementCost,
                blockade: blockadeCost,
                floodworks: floodworksCost,
            },
        });

        onBuild(action, x, y);
        selectedBuildAction = null;
    }

    function selectBuildAction(action: TargetBuildAction) {
        selectedBuildAction = selectedBuildAction === action ? null : action;

        debugLog("build.select", {
            action,
            selectedBuildAction,
            playerId,
            currentPlayer: game.CurrentPlayer,
            currentPhase: game.CurrentPhase,
            isMyTurn,
        });
    }

    function useSelectedInstant(targetPlayerId = 0) {
        if (selectedHandIndex < 0 || !selectedHandItem) return;

        onUseDraft(selectedHandIndex, 0, 0, targetPlayerId);
    }
</script>

<main
    class="relative min-h-screen overflow-hidden bg-[#17313a] font-sans text-[#f8efe0]"
>
    <div class="pointer-events-none absolute inset-0 bg-[#15323a]">
        <div class="absolute inset-0 bg-board-flat"></div>
        <div class="absolute inset-0 bg-board-texture opacity-[0.16]"></div>
        <div
            class="absolute inset-x-0 bottom-0 h-40 bg-gradient-to-t from-black/18 to-transparent"
        ></div>
    </div>

    <header
        class="relative z-10 mx-auto flex max-w-7xl items-center justify-between px-6 py-5 lg:px-12"
    >
        <div class="flex items-center gap-3">
            <div
                class="grid h-11 w-11 place-items-center rounded-2xl bg-[#f2c36b] text-xl font-black text-[#142833] shadow-[0_8px_0_rgba(0,0,0,0.16)] ring-1 ring-white/20"
            >
                <span class="logo-diamond">◈</span>
            </div>

            <div>
                <div
                    class="text-xl font-semibold tracking-tight text-[#fff7e8]"
                >
                    Frontiers
                </div>
                <div
                    class="text-xs font-semibold uppercase tracking-[0.22em] text-[#9fc9c5]"
                >
                    Room {roomId}
                </div>
            </div>
        </div>

        <div class="flex items-center gap-3">
            <button
                class="cursor-pointer rounded-xl bg-[#f8efe0]/10 px-3 py-2 text-sm font-bold text-[#fff7e8] ring-1 ring-[#f8efe0]/20 hover:bg-[#f8efe0]/16"
                type="button"
                on:click={onCopyRoomCode}
            >
                Copy Code
            </button>

            <button
                class="cursor-pointer rounded-xl bg-[#b94b3f] px-4 py-2 text-sm font-bold text-white shadow-sm transition hover:bg-[#c9574a]"
                type="button"
                on:click={onLeaveRoom}
            >
                Leave
            </button>
        </div>
    </header>

    <section
        class="relative z-10 mx-auto grid w-full max-w-[1640px] gap-5 px-4 pb-10 pt-3 lg:grid-cols-[310px_minmax(0,1fr)_350px] lg:px-6 xl:grid-cols-[330px_minmax(0,1fr)_360px]"
    >
        <aside
            class="min-w-0 space-y-5 lg:max-h-[calc(100vh-112px)] lg:overflow-y-auto lg:pr-1"
        >
            <section
                class="rounded-3xl bg-[#23444c] p-5 shadow-md ring-1 ring-[#f8efe0]/10"
            >
                <div
                    class="text-sm font-black uppercase tracking-[0.22em] text-[#9fc9c5]"
                >
                    Match
                </div>

                <div class="mt-4 grid grid-cols-3 gap-2">
                    <div
                        class="rounded-2xl bg-[#f8efe0]/10 p-3 text-center ring-1 ring-[#f8efe0]/10"
                    >
                        <div
                            class="text-[10px] font-black uppercase tracking-wider text-[#9fc9c5]"
                        >
                            Round
                        </div>
                        <div class="mt-1 text-2xl font-black text-[#fff7e8]">
                            {game.Round}
                        </div>
                    </div>

                    <div
                        class="rounded-2xl bg-[#f2c36b] p-3 text-center text-[#142833]"
                    >
                        <div
                            class="text-[10px] font-black uppercase tracking-wider opacity-70"
                        >
                            Phase
                        </div>
                        <div class="mt-1 text-sm font-black">
                            {currentPhaseName}
                        </div>
                    </div>

                    <div
                        class={[
                            "rounded-2xl p-3 text-center ring-1",
                            isMyTurn
                                ? "bg-[#73c4bd] text-[#102b38] ring-[#73c4bd]"
                                : "bg-[#f8efe0]/10 text-[#fff7e8] ring-[#f8efe0]/10",
                        ].join(" ")}
                    >
                        <div
                            class="text-[10px] font-black uppercase tracking-wider opacity-70"
                        >
                            Turn
                        </div>
                        <div class="mt-1 text-sm font-black">
                            {currentPlayerName}
                        </div>
                    </div>
                </div>
            </section>

            <section
                class="rounded-3xl bg-[#23444c] p-5 shadow-md ring-1 ring-[#f8efe0]/10"
            >
                <h2 class="text-xl font-black text-[#fff7e8]">Players</h2>

                <div class="mt-4 space-y-3">
                    {#each game.Players as player}
                        <div
                            class={[
                                "rounded-2xl p-4 ring-1",
                                player.Id === game.CurrentPlayer
                                    ? "bg-[#f2c36b]/16 ring-[#f2c36b]/45"
                                    : "bg-[#f8efe0]/10 ring-[#f8efe0]/10",
                            ].join(" ")}
                        >
                            <div class="flex items-center gap-3">
                                <div
                                    class={[
                                        "grid h-10 w-10 place-items-center rounded-2xl text-sm font-black text-white",
                                        playerColor(player.Id),
                                    ].join(" ")}
                                >
                                    P{player.Id}
                                </div>

                                <div>
                                    <div class="font-black text-[#fff7e8]">
                                        {player.Id === playerId
                                            ? "You"
                                            : `Player ${player.Id}`}
                                    </div>

                                    <div
                                        class="mt-1 text-xs font-semibold text-[#9fc9c5]"
                                    >
                                        {player.Id === game.CurrentPlayer
                                            ? "Taking turn"
                                            : "Waiting"}
                                    </div>
                                </div>
                            </div>

                            <div class="mt-3 flex flex-wrap gap-2">
                                {#each handItems(player) as card}
                                    <MiniHandCard item={card} />
                                {/each}

                                {#if handItems(player).length === 0}
                                    <MiniHandCard item={null} />
                                {/if}
                            </div>

                            <div class="mt-3 flex flex-nowrap gap-1.5">
                                <ResourceIcon
                                    resource="wood"
                                    amount={resourceAmount(player, 1)}
                                    pulse={resourcePulse(player.Id, 1)}
                                />

                                <ResourceIcon
                                    resource="stone"
                                    amount={resourceAmount(player, 2)}
                                    pulse={resourcePulse(player.Id, 2)}
                                />

                                <ResourceIcon
                                    resource="grain"
                                    amount={resourceAmount(player, 3)}
                                    pulse={resourcePulse(player.Id, 3)}
                                />

                                <ResourceIcon
                                    resource="relic"
                                    amount={resourceAmount(player, 4)}
                                    pulse={resourcePulse(player.Id, 4)}
                                />

                                {#if player.FloodTokens && player.FloodTokens > 0}
                                    <div
                                        class="inline-flex h-8 min-w-12 items-center justify-center gap-1.5 rounded-xl border-2 border-[#327b8d] bg-[#6eb8c5] px-2 text-xs font-black text-[#102b38] shadow-sm"
                                        title="Flood tokens"
                                    >
                                        ≈ {player.FloodTokens}
                                    </div>
                                {/if}
                            </div>
                        </div>
                    {/each}
                </div>
            </section>
        </aside>

        <div class="min-h-0 min-w-0 lg:sticky lg:top-24 lg:self-start">
            <Board
                {game}
                {playerId}
                {role}
                {selectedBuildAction}
                {selectedHandIndex}
                {selectedHandItem}
                {onPlaceTile}
                {onUseDraft}
                onBuild={handleBuild}
            />
        </div>

        <aside
            class="min-w-0 space-y-5 lg:max-h-[calc(100vh-112px)] lg:overflow-y-auto lg:pr-1"
        >
            {#if hasMyResourceGain}
                <section
                    class="resource-toast rounded-3xl bg-[#73c4bd] p-4 text-[#102b38] shadow-md ring-1 ring-white/20"
                >
                    <div
                        class="text-xs font-black uppercase tracking-wider opacity-70"
                    >
                        Gained
                    </div>

                    <div class="mt-2 flex flex-wrap gap-2">
                        {#if myWoodGain > 0}
                            <ResourceIcon
                                resource="wood"
                                amount={`+${myWoodGain}`}
                                size="md"
                                pulse
                            />
                        {/if}

                        {#if myStoneGain > 0}
                            <ResourceIcon
                                resource="stone"
                                amount={`+${myStoneGain}`}
                                size="md"
                                pulse
                            />
                        {/if}

                        {#if myGrainGain > 0}
                            <ResourceIcon
                                resource="grain"
                                amount={`+${myGrainGain}`}
                                size="md"
                                pulse
                            />
                        {/if}

                        {#if myRelicGain > 0}
                            <ResourceIcon
                                resource="relic"
                                amount={`+${myRelicGain}`}
                                size="md"
                                pulse
                            />
                        {/if}
                    </div>
                </section>
            {/if}

            {#if isMyTurn && game.CurrentPhase === GamePhase.Place}
                <section
                    class="rounded-3xl bg-[#23444c] p-5 shadow-md ring-1 ring-[#f8efe0]/10"
                >
                    <div class="flex items-center justify-between gap-3">
                        <h2 class="text-xl font-black text-[#fff7e8]">Hand</h2>

                        <div
                            class="rounded-xl bg-[#73c4bd] px-3 py-1 text-xs font-black uppercase tracking-wider text-[#102b38]"
                        >
                            Your turn
                        </div>
                    </div>

                    {#if myHand.length > 0}
                        <div class="mt-4 grid gap-3">
                            {#each myHand as card, index}
                                <CardTooltip
                                    item={card}
                                    hint={selectedHandIndex === index
                                        ? "Selected"
                                        : "Click to select"}
                                >
                                    <button
                                        class={[
                                            "cursor-pointer rounded-2xl text-left transition hover:-translate-y-0.5",
                                            selectedHandIndex === index
                                                ? "ring-4 ring-[#f2c36b]"
                                                : "ring-1 ring-transparent",
                                        ].join(" ")}
                                        type="button"
                                        on:click={() =>
                                            (selectedHandIndex = index)}
                                    >
                                        <HandCard item={card} size="md" />
                                    </button>
                                </CardTooltip>
                            {/each}
                        </div>

                        {#if selectedHandItem?.Kind === DraftKind.Action && selectedHandItem.Action === Action.Expansion}
                            <button
                                class="mt-4 w-full cursor-pointer rounded-2xl bg-[#73c4bd] px-5 py-3 font-black text-[#102b38] shadow-[0_6px_0_rgba(0,0,0,0.18)] transition hover:bg-[#85d8d1] active:translate-y-1"
                                type="button"
                                on:click={() => useSelectedInstant()}
                            >
                                Use Expansion
                            </button>
                        {/if}

                        {#if selectedHandItem?.Kind === DraftKind.Action && selectedHandItem.Action === Action.Raid}
                            <div class="mt-4 grid gap-2">
                                {#each game.Players.filter((p) => p.Id !== playerId) as target}
                                    <button
                                        class="cursor-pointer rounded-2xl bg-[#b94b3f] px-5 py-3 font-black text-white shadow-[0_6px_0_rgba(0,0,0,0.18)] transition hover:bg-[#c9574a] active:translate-y-1"
                                        type="button"
                                        on:click={() =>
                                            useSelectedInstant(target.Id)}
                                    >
                                        Raid P{target.Id}
                                    </button>
                                {/each}
                            </div>
                        {/if}

                        {#if canDiscardSelected}
                            <button
                                class="mt-4 w-full cursor-pointer rounded-2xl bg-[#b94b3f] px-5 py-3 font-black text-white shadow-[0_6px_0_rgba(0,0,0,0.18)] transition hover:bg-[#c9574a] active:translate-y-1"
                                type="button"
                                on:click={() =>
                                    onDiscardDraft(selectedHandIndex)}
                            >
                                Discard
                            </button>
                        {/if}
                    {:else}
                        <button
                            class="mt-4 w-full cursor-pointer rounded-2xl bg-[#f8efe0]/10 px-5 py-3 font-black text-[#fff7e8] shadow-[0_6px_0_rgba(0,0,0,0.18)] ring-1 ring-[#f8efe0]/20 transition hover:bg-[#f8efe0]/16 active:translate-y-1"
                            type="button"
                            on:click={onPassPlace}
                        >
                            Continue
                        </button>
                    {/if}
                </section>
            {/if}

            {#if isMyTurn && game.CurrentPhase === GamePhase.Build}
                <section
                    class="rounded-3xl bg-[#23444c] p-5 shadow-md ring-1 ring-[#f8efe0]/10"
                >
                    <div class="flex items-center justify-between gap-3">
                        <h2 class="text-xl font-black text-[#fff7e8]">Build</h2>

                        <div
                            class="rounded-xl bg-[#73c4bd] px-3 py-1 text-xs font-black uppercase tracking-wider text-[#102b38]"
                        >
                            Your turn
                        </div>
                    </div>

                    {#if selectedBuildAction}
                        <div
                            class="mt-4 rounded-2xl bg-[#f2c36b]/20 p-3 text-center text-sm font-black text-[#f8efe0] ring-1 ring-[#f2c36b]/40"
                        >
                            {selectedBuildAction === "city"
                                ? "Click your outpost to upgrade it"
                                : selectedBuildAction === "settlement"
                                  ? "Click friendly empty land"
                                  : selectedBuildAction === "blockade"
                                    ? "Click enemy land or adjacent neutral land"
                                    : selectedBuildAction === "flood"
                                      ? "Click a tile without structure"
                                      : "Click a valid land tile"}
                        </div>
                    {/if}

                    <div class="mt-4 grid grid-cols-2 gap-3">
                        <BuildActionTooltip
                            action="outpost"
                            hint={costTitle(
                                outpostCost,
                                canAffordOutpost,
                                hasOutpostTarget
                                    ? "Build Outpost"
                                    : "No valid target",
                            )}
                        >
                            <button
                                class={[
                                    "h-full w-full rounded-2xl p-4 text-center font-black shadow-[0_6px_0_rgba(0,0,0,0.18)] transition active:translate-y-1",
                                    selectedBuildAction === "outpost"
                                        ? "bg-[#f2c36b] text-[#142833]"
                                        : "bg-[#f8efe0]/10 text-[#fff7e8] ring-1 ring-[#f8efe0]/20",
                                    canSelectOutpost
                                        ? "cursor-pointer hover:bg-[#f8efe0]/16"
                                        : "cursor-not-allowed opacity-45",
                                ].join(" ")}
                                type="button"
                                disabled={!canSelectOutpost}
                                title={costTitle(
                                    outpostCost,
                                    canAffordOutpost,
                                    hasOutpostTarget
                                        ? "Build Outpost"
                                        : "No valid target",
                                )}
                                on:click={() => selectBuildAction("outpost")}
                            >
                                <div class="text-3xl">⌂</div>
                                <div
                                    class="mt-1 text-xs uppercase tracking-wider"
                                >
                                    Outpost
                                </div>
                                <div class="mt-2">
                                    <CostBadge
                                        wood={outpostCost.wood ?? 0}
                                        stone={outpostCost.stone ?? 0}
                                        grain={outpostCost.grain ?? 0}
                                        relic={outpostCost.relic ?? 0}
                                        affordable={canAffordOutpost}
                                    />
                                </div>
                            </button>
                        </BuildActionTooltip>

                        <BuildActionTooltip
                            action="settlement"
                            hint={costTitle(
                                settlementCost,
                                canAffordSettlement,
                                hasSettlementTarget
                                    ? "Build Settlement"
                                    : "Needs friendly empty land",
                            )}
                        >
                            <button
                                class={[
                                    "h-full w-full rounded-2xl p-4 text-center font-black shadow-[0_6px_0_rgba(0,0,0,0.18)] transition active:translate-y-1",
                                    selectedBuildAction === "settlement"
                                        ? "bg-[#f2c36b] text-[#142833]"
                                        : "bg-[#f8efe0]/10 text-[#fff7e8] ring-1 ring-[#f8efe0]/20",
                                    canSelectSettlement
                                        ? "cursor-pointer hover:bg-[#f8efe0]/16"
                                        : "cursor-not-allowed opacity-45",
                                ].join(" ")}
                                type="button"
                                disabled={!canSelectSettlement}
                                title={costTitle(
                                    settlementCost,
                                    canAffordSettlement,
                                    hasSettlementTarget
                                        ? "Build Settlement"
                                        : "Needs friendly empty land",
                                )}
                                on:click={() => selectBuildAction("settlement")}
                            >
                                <div class="text-3xl">◈</div>
                                <div
                                    class="mt-1 text-xs uppercase tracking-wider"
                                >
                                    Settlement
                                </div>
                                <div class="mt-2">
                                    <CostBadge
                                        wood={settlementCost.wood ?? 0}
                                        stone={settlementCost.stone ?? 0}
                                        grain={settlementCost.grain ?? 0}
                                        relic={settlementCost.relic ?? 0}
                                        affordable={canAffordSettlement}
                                    />
                                </div>
                            </button>
                        </BuildActionTooltip>

                        <BuildActionTooltip
                            action="city"
                            hint={costTitle(
                                cityCost,
                                canAffordCity,
                                hasCityTarget
                                    ? "Upgrade Outpost to City"
                                    : "Requires your outpost",
                            )}
                        >
                            <button
                                class={[
                                    "h-full w-full rounded-2xl p-4 text-center font-black shadow-[0_6px_0_rgba(0,0,0,0.18)] transition active:translate-y-1",
                                    selectedBuildAction === "city"
                                        ? "bg-[#f2c36b] text-[#142833]"
                                        : "bg-[#f8efe0]/10 text-[#fff7e8] ring-1 ring-[#f8efe0]/20",
                                    canSelectCity
                                        ? "cursor-pointer hover:bg-[#f8efe0]/16"
                                        : "cursor-not-allowed opacity-45",
                                ].join(" ")}
                                type="button"
                                disabled={!canSelectCity}
                                title={costTitle(
                                    cityCost,
                                    canAffordCity,
                                    hasCityTarget
                                        ? "Upgrade Outpost to City"
                                        : "Requires your outpost",
                                )}
                                on:click={() => selectBuildAction("city")}
                            >
                                <div class="text-3xl">▦</div>
                                <div
                                    class="mt-1 text-xs uppercase tracking-wider"
                                >
                                    City
                                </div>
                                <div class="mt-2">
                                    <CostBadge
                                        wood={cityCost.wood ?? 0}
                                        stone={cityCost.stone ?? 0}
                                        grain={cityCost.grain ?? 0}
                                        relic={cityCost.relic ?? 0}
                                        affordable={canAffordCity}
                                    />
                                </div>
                            </button>
                        </BuildActionTooltip>

                        <BuildActionTooltip
                            action="blockade"
                            hint={costTitle(
                                blockadeCost,
                                canAffordBlockade,
                                hasBlockadeTarget
                                    ? "Build Blockade"
                                    : "No valid blockade target",
                            )}
                        >
                            <button
                                class={[
                                    "h-full w-full rounded-2xl p-4 text-center font-black shadow-[0_6px_0_rgba(0,0,0,0.18)] transition active:translate-y-1",
                                    selectedBuildAction === "blockade"
                                        ? "bg-[#f2c36b] text-[#142833]"
                                        : "bg-[#f8efe0]/10 text-[#fff7e8] ring-1 ring-[#f8efe0]/20",
                                    canSelectBlockade
                                        ? "cursor-pointer hover:bg-[#f8efe0]/16"
                                        : "cursor-not-allowed opacity-45",
                                ].join(" ")}
                                type="button"
                                disabled={!canSelectBlockade}
                                title={costTitle(
                                    blockadeCost,
                                    canAffordBlockade,
                                    hasBlockadeTarget
                                        ? "Build Blockade"
                                        : "No valid blockade target",
                                )}
                                on:click={() => selectBuildAction("blockade")}
                            >
                                <div class="text-3xl">✕</div>
                                <div
                                    class="mt-1 text-xs uppercase tracking-wider"
                                >
                                    Blockade
                                </div>
                                <div class="mt-2">
                                    <CostBadge
                                        wood={blockadeCost.wood ?? 0}
                                        stone={blockadeCost.stone ?? 0}
                                        grain={blockadeCost.grain ?? 0}
                                        relic={blockadeCost.relic ?? 0}
                                        affordable={canAffordBlockade}
                                    />
                                </div>
                            </button>
                        </BuildActionTooltip>
                    </div>

                    <div class="mt-3 grid grid-cols-2 gap-3">
                        <BuildActionTooltip
                            action="floodworks"
                            hint={costTitle(
                                floodworksCost,
                                canAffordFloodworks,
                                "Buy 3 flood tokens",
                            )}
                        >
                            <button
                                class={[
                                    "h-full w-full rounded-2xl p-4 text-center font-black shadow-[0_6px_0_rgba(0,0,0,0.18)] transition active:translate-y-1",
                                    canBuyFloodworks
                                        ? "cursor-pointer bg-[#6eb8c5] text-[#102b38] hover:bg-[#85d8d1]"
                                        : "cursor-not-allowed bg-[#f8efe0]/10 text-[#fff7e8] opacity-45 ring-1 ring-[#f8efe0]/20",
                                ].join(" ")}
                                type="button"
                                disabled={!canBuyFloodworks}
                                title={costTitle(
                                    floodworksCost,
                                    canAffordFloodworks,
                                    "Buy 3 flood tokens",
                                )}
                                on:click={() => onBuild("floodworks", 0, 0)}
                            >
                                <div class="text-3xl">≈</div>
                                <div
                                    class="mt-1 text-xs uppercase tracking-wider"
                                >
                                    Floodworks
                                </div>
                                <div class="mt-2">
                                    <CostBadge
                                        wood={floodworksCost.wood ?? 0}
                                        stone={floodworksCost.stone ?? 0}
                                        grain={floodworksCost.grain ?? 0}
                                        relic={floodworksCost.relic ?? 0}
                                        affordable={canAffordFloodworks}
                                    />
                                </div>
                            </button>
                        </BuildActionTooltip>

                        <BuildActionTooltip
                            action="flood"
                            hint={canUseFloodToken
                                ? "Convert one tile to river"
                                : "Need flood token and valid target"}
                        >
                            <button
                                class={[
                                    "h-full w-full rounded-2xl p-4 text-center font-black shadow-[0_6px_0_rgba(0,0,0,0.18)] transition active:translate-y-1",
                                    selectedBuildAction === "flood"
                                        ? "bg-[#f2c36b] text-[#142833]"
                                        : "bg-[#f8efe0]/10 text-[#fff7e8] ring-1 ring-[#f8efe0]/20",
                                    canUseFloodToken
                                        ? "cursor-pointer hover:bg-[#f8efe0]/16"
                                        : "cursor-not-allowed opacity-45",
                                ].join(" ")}
                                type="button"
                                disabled={!canUseFloodToken}
                                title={canUseFloodToken
                                    ? "Convert one tile to river"
                                    : "Need flood token and valid target"}
                                on:click={() => selectBuildAction("flood")}
                            >
                                <div class="text-3xl">≈</div>
                                <div
                                    class="mt-1 text-xs uppercase tracking-wider"
                                >
                                    Flood
                                </div>
                                <div class="mt-2 text-xs font-black">
                                    Tokens: {me?.FloodTokens ?? 0}
                                </div>
                            </button>
                        </BuildActionTooltip>
                    </div>

                    <button
                        class="mt-4 w-full cursor-pointer rounded-2xl bg-[#f8efe0]/10 px-5 py-3 font-black text-[#fff7e8] shadow-[0_6px_0_rgba(0,0,0,0.18)] ring-1 ring-[#f8efe0]/20 transition hover:bg-[#f8efe0]/16 active:translate-y-1"
                        type="button"
                        on:click={onPassBuild}
                    >
                        Pass
                    </button>
                </section>
            {/if}

            {#if error}
                <section
                    class="rounded-3xl bg-[#23444c] p-5 shadow-md ring-1 ring-[#f8efe0]/10"
                >
                    <div
                        class="rounded-2xl bg-[#b94b3f] px-5 py-3 text-sm font-semibold text-white"
                    >
                        {error}
                    </div>
                </section>
            {/if}

            <Market {game} {playerId} {role} {onPick} />
        </aside>
    </section>
</main>

<style>
    .logo-diamond {
        display: block;
        line-height: 1;
        transform: translateY(-1px);
    }

    .bg-board-flat {
        background: linear-gradient(180deg, #173943 0%, #102832 100%);
    }

    .bg-board-texture {
        background-image: radial-gradient(
            circle,
            rgba(255, 255, 255, 0.055) 1px,
            transparent 1px
        );
        background-size: 28px 28px;
    }

    .resource-toast {
        animation: toast-pop 720ms ease-out;
    }

    @keyframes toast-pop {
        0% {
            transform: translateY(-8px);
            opacity: 0;
        }

        100% {
            transform: translateY(0);
            opacity: 1;
        }
    }
</style>
