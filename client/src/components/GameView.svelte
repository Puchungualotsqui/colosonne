<script lang="ts">
    import { tick } from "svelte";

    import ResourceIcon from "./ResourceIcon.svelte";
    import ResourceFlightLayer, {
        type ResourceFlight,
    } from "./ResourceFlightLayer.svelte";
    import Board from "./Board.svelte";
    import Market from "./Market.svelte";
    import HandCard from "./HandCard.svelte";
    import MiniHandCard from "./MiniHandCard.svelte";
    import CardTooltip from "./CardTooltip.svelte";
    import BuildButton from "./BuildButton.svelte";

    import AppShell from "./ui/AppShell.svelte";
    import AppHeader from "./ui/AppHeader.svelte";
    import Panel from "./ui/Panel.svelte";

    import { canUseHandItem, hasBuildTarget } from "../lib/rules";
    import { ui } from "../lib/uiClasses";
    import { debugLog } from "../lib/debug";

    import {
        Action,
        DraftKind,
        GamePhase,
        Resource,
        Structure,
        type BuildAction,
        type BuildCostsByPlayer,
        type DraftItem,
        type GameEvent,
        type GameState,
        type Player,
        type ResourceCostResponse,
        type ScoresByPlayer,
        type TargetBuildAction,
    } from "../lib/types";

    export let game: GameState;
    export let roomId = "";
    export let playerId = 0;
    export let role: "player" | "spectator" | "" = "";
    export let error = "";
    export let buildCosts: BuildCostsByPlayer = {};
    export let scores: ScoresByPlayer = {};
    export let events: GameEvent[] = [];

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

    type Point = {
        x: number;
        y: number;
    };

    type BoardHandle = {
        getTileCenterViewport: (x: number, y: number) => Point | null;
    };

    let boardRef: BoardHandle | null = null;
    let resourceAnchorEls: Record<string, HTMLElement | undefined> = {};

    let resourceFlights: ResourceFlight[] = [];
    let processedEventIds = new Set<number>();
    let flightSeq = 0;
    let eventScheduleSeq = 0;

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

    $: selectedHandIsUsable = canUseHandItem(game, playerId, selectedHandItem);
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

    $: hasOutpostTarget = hasBuildTarget(game, playerId, "outpost");
    $: hasCityTarget = hasBuildTarget(game, playerId, "city");
    $: hasSettlementTarget = hasBuildTarget(game, playerId, "settlement");
    $: hasBlockadeTarget = hasBuildTarget(game, playerId, "blockade");
    $: hasFloodTarget = hasBuildTarget(game, playerId, "flood");

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

    $: if (events.length > 0) {
        scheduleEventAnimations(events);
    }

    $: marketCardsRemaining = game.Market.filter(
        (item) => item !== null,
    ).length;

    $: handCardsRemaining = game.Players.reduce(
        (total, player) => total + (player.Hand?.length ?? 0),
        0,
    );

    $: tileCardsRemaining =
        game.Deck.filter((item) => item.Kind === DraftKind.Tile).length +
        game.Market.filter((item) => item?.Kind === DraftKind.Tile).length +
        game.Players.reduce(
            (total, player) =>
                total +
                (player.Hand?.filter((item) => item.Kind === DraftKind.Tile)
                    .length ?? 0),
            0,
        );

    function scoreForPlayer(targetPlayerId: number) {
        return scores[String(targetPlayerId)] ?? 0;
    }

    async function scheduleEventAnimations(nextEvents: GameEvent[]) {
        const pending = nextEvents.filter(
            (event) => !processedEventIds.has(event.id),
        );

        if (pending.length === 0) return;

        for (const event of pending) {
            processedEventIds.add(event.id);
        }

        const scheduleId = ++eventScheduleSeq;

        await tick();

        if (scheduleId !== eventScheduleSeq) {
            // A newer event batch arrived. Still safe to animate pending events,
            // because IDs are already de-duplicated.
        }

        for (const event of pending) {
            enqueueEventAnimation(event);
        }
    }

    function enqueueEventAnimation(event: GameEvent) {
        switch (event.kind) {
            case "resource_gain":
                enqueueResourceGain(event);
                return;

            case "resource_transfer":
                enqueueResourceTransfer(event);
                return;

            default:
                return;
        }
    }

    function enqueueResourceGain(event: GameEvent) {
        if (!event.toPlayer || !event.resource || !event.amount) return;

        const to = getResourceAnchorCenter(event.toPlayer, event.resource);
        if (!to) return;

        const from = event.from
            ? boardRef?.getTileCenterViewport(event.from.x, event.from.y)
            : actionSourcePoint(event.toPlayer, event.resource);

        if (!from) return;

        pushResourceFlight({
            resource: event.resource,
            amount: event.amount,
            from,
            to,
            label: `+${event.amount}`,
        });
    }

    function enqueueResourceTransfer(event: GameEvent) {
        if (
            !event.fromPlayer ||
            !event.toPlayer ||
            !event.resource ||
            !event.amount
        ) {
            return;
        }

        const from =
            getResourceAnchorCenter(event.fromPlayer, event.resource) ??
            actionSourcePoint(event.fromPlayer, event.resource);

        const to = getResourceAnchorCenter(event.toPlayer, event.resource);

        if (!from || !to) return;

        pushResourceFlight({
            resource: event.resource,
            amount: event.amount,
            from,
            to,
            label: `+${event.amount}`,
        });
    }

    function pushResourceFlight(input: {
        resource: Resource;
        amount: number;
        from: Point;
        to: Point;
        label?: string;
    }) {
        const id = ++flightSeq;
        const jitter = ((id % 5) - 2) * 7;

        const flight: ResourceFlight = {
            id,
            resource: input.resource,
            amount: input.amount,
            fromX: input.from.x + jitter,
            fromY: input.from.y,
            toX: input.to.x + jitter * 0.25,
            toY: input.to.y,
            label: input.label,
        };

        resourceFlights = [...resourceFlights, flight];

        window.setTimeout(() => {
            resourceFlights = resourceFlights.filter((item) => item.id !== id);
        }, 1700);
    }

    function getResourceAnchorCenter(
        targetPlayerId: number,
        resource: Resource,
    ): Point | null {
        const key = resourceAnchorKey(targetPlayerId, resource);
        const element = resourceAnchorEls[key];

        if (!element) return null;

        const rect = element.getBoundingClientRect();

        return {
            x: rect.left + rect.width / 2,
            y: rect.top + rect.height / 2,
        };
    }

    function actionSourcePoint(
        targetPlayerId: number,
        resource: Resource,
    ): Point | null {
        const target = getResourceAnchorCenter(targetPlayerId, resource);

        if (target) {
            return {
                x: target.x,
                y: target.y - 120,
            };
        }

        return {
            x: window.innerWidth / 2,
            y: window.innerHeight / 2,
        };
    }

    function resourceAnchorKey(targetPlayerId: number, resource: Resource) {
        return `${targetPlayerId}:${resource}`;
    }

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

    function handleInstantBuild(action: BuildAction) {
        onBuild(action, 0, 0);
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
        const usableIndex = myHand.findIndex((item) =>
            canUseHandItem(game, playerId, item),
        );

        if (usableIndex >= 0) return usableIndex;
        return myHand.length > 0 ? 0 : -1;
    }

    function playerColor(targetPlayerId: number) {
        if (targetPlayerId === 1) return "bg-[#1d4e89]";
        if (targetPlayerId === 2) return "bg-[#b94b3f]";
        return "bg-[#6b4a2f]";
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

    function selectedBuildHint(action: TargetBuildAction) {
        switch (action) {
            case "city":
                return "Click your outpost to upgrade it";

            case "settlement":
                return "Click friendly empty land";

            case "blockade":
                return "Click enemy land or adjacent neutral land";

            case "flood":
                return "Click a tile without structure";

            case "outpost":
                return "Click neutral or friendly empty land";

            default:
                return "Click a valid target";
        }
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

<AppShell>
    <AppHeader subtitle={`Room ${roomId}`}>
        <svelte:fragment slot="actions">
            <button
                class={ui.button.smallSecondary}
                type="button"
                on:click={onCopyRoomCode}
            >
                Copy Code
            </button>

            <button
                class={ui.button.danger}
                type="button"
                on:click={onLeaveRoom}
            >
                Leave
            </button>
        </svelte:fragment>
    </AppHeader>

    <section
        class="relative z-10 mx-auto grid w-full max-w-[1640px] gap-5 px-4 pb-10 pt-3 lg:grid-cols-[310px_minmax(0,1fr)_350px] lg:px-6 xl:grid-cols-[330px_minmax(0,1fr)_360px]"
    >
        <aside
            class="min-w-0 space-y-5 lg:max-h-[calc(100vh-112px)] lg:overflow-y-auto lg:pr-1"
        >
            <Panel>
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

                    <div
                        class="col-span-3 rounded-2xl bg-[#f8efe0]/10 p-3 text-center ring-1 ring-[#f8efe0]/10"
                        title="Tile cards still left to be placed from deck, market, and hands"
                    >
                        <div
                            class="text-[10px] font-black uppercase tracking-wider text-[#9fc9c5]"
                        >
                            Tiles left to place
                        </div>
                        <div class="mt-1 text-3xl font-black text-[#fff7e8]">
                            {tileCardsRemaining}
                        </div>
                        <div
                            class="mt-1 text-[11px] font-semibold text-[#9fc9c5]"
                        >
                            Deck + market + hands
                        </div>
                    </div>
                </div>
            </Panel>

            <Panel>
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

                                <div class="min-w-0">
                                    <div
                                        class="flex flex-wrap items-center gap-2"
                                    >
                                        <div class="font-black text-[#fff7e8]">
                                            {player.Id === playerId
                                                ? "You"
                                                : `Player ${player.Id}`}
                                        </div>

                                        <div
                                            class="rounded-lg bg-[#f2c36b] px-2 py-0.5 text-[10px] font-black text-[#142833]"
                                            title="Current victory points"
                                        >
                                            {scoreForPlayer(player.Id)} VP
                                        </div>
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
                                <span
                                    class="inline-flex"
                                    bind:this={
                                        resourceAnchorEls[
                                            `${player.Id}:${RES_WOOD}`
                                        ]
                                    }
                                >
                                    <ResourceIcon
                                        resource="wood"
                                        amount={resourceAmount(
                                            player,
                                            RES_WOOD,
                                        )}
                                        pulse={resourcePulse(
                                            player.Id,
                                            RES_WOOD,
                                        )}
                                    />
                                </span>

                                <span
                                    class="inline-flex"
                                    bind:this={
                                        resourceAnchorEls[
                                            `${player.Id}:${RES_STONE}`
                                        ]
                                    }
                                >
                                    <ResourceIcon
                                        resource="stone"
                                        amount={resourceAmount(
                                            player,
                                            RES_STONE,
                                        )}
                                        pulse={resourcePulse(
                                            player.Id,
                                            RES_STONE,
                                        )}
                                    />
                                </span>

                                <span
                                    class="inline-flex"
                                    bind:this={
                                        resourceAnchorEls[
                                            `${player.Id}:${RES_GRAIN}`
                                        ]
                                    }
                                >
                                    <ResourceIcon
                                        resource="grain"
                                        amount={resourceAmount(
                                            player,
                                            RES_GRAIN,
                                        )}
                                        pulse={resourcePulse(
                                            player.Id,
                                            RES_GRAIN,
                                        )}
                                    />
                                </span>

                                <span
                                    class="inline-flex"
                                    bind:this={
                                        resourceAnchorEls[
                                            `${player.Id}:${RES_RELIC}`
                                        ]
                                    }
                                >
                                    <ResourceIcon
                                        resource="relic"
                                        amount={resourceAmount(
                                            player,
                                            RES_RELIC,
                                        )}
                                        pulse={resourcePulse(
                                            player.Id,
                                            RES_RELIC,
                                        )}
                                    />
                                </span>

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
            </Panel>
        </aside>

        <div class="min-h-0 min-w-0 lg:sticky lg:top-24 lg:self-start">
            <Board
                bind:this={boardRef}
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
                <Panel>
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
                </Panel>
            {/if}

            {#if isMyTurn && game.CurrentPhase === GamePhase.Build}
                <Panel>
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
                            {selectedBuildHint(selectedBuildAction)}
                        </div>
                    {/if}

                    <div class="mt-4 grid grid-cols-2 gap-3">
                        <BuildButton
                            action="outpost"
                            label="Outpost"
                            icon="⌂"
                            cost={outpostCost}
                            affordable={canAffordOutpost}
                            enabled={canSelectOutpost}
                            selected={selectedBuildAction === "outpost"}
                            title={costTitle(
                                outpostCost,
                                canAffordOutpost,
                                hasOutpostTarget
                                    ? "Build Outpost"
                                    : "No valid target",
                            )}
                            hint={costTitle(
                                outpostCost,
                                canAffordOutpost,
                                hasOutpostTarget
                                    ? "Build Outpost"
                                    : "No valid target",
                            )}
                            onSelect={selectBuildAction}
                        />

                        <BuildButton
                            action="settlement"
                            label="Settlement"
                            icon="◈"
                            cost={settlementCost}
                            affordable={canAffordSettlement}
                            enabled={canSelectSettlement}
                            selected={selectedBuildAction === "settlement"}
                            title={costTitle(
                                settlementCost,
                                canAffordSettlement,
                                hasSettlementTarget
                                    ? "Build Settlement"
                                    : "Needs friendly empty land",
                            )}
                            hint={costTitle(
                                settlementCost,
                                canAffordSettlement,
                                hasSettlementTarget
                                    ? "Build Settlement"
                                    : "Needs friendly empty land",
                            )}
                            onSelect={selectBuildAction}
                        />

                        <BuildButton
                            action="city"
                            label="City"
                            icon="▦"
                            cost={cityCost}
                            affordable={canAffordCity}
                            enabled={canSelectCity}
                            selected={selectedBuildAction === "city"}
                            title={costTitle(
                                cityCost,
                                canAffordCity,
                                hasCityTarget
                                    ? "Upgrade Outpost to City"
                                    : "Requires your outpost",
                            )}
                            hint={costTitle(
                                cityCost,
                                canAffordCity,
                                hasCityTarget
                                    ? "Upgrade Outpost to City"
                                    : "Requires your outpost",
                            )}
                            onSelect={selectBuildAction}
                        />

                        <BuildButton
                            action="blockade"
                            label="Blockade"
                            icon="✕"
                            cost={blockadeCost}
                            affordable={canAffordBlockade}
                            enabled={canSelectBlockade}
                            selected={selectedBuildAction === "blockade"}
                            title={costTitle(
                                blockadeCost,
                                canAffordBlockade,
                                hasBlockadeTarget
                                    ? "Build Blockade"
                                    : "No valid blockade target",
                            )}
                            hint={costTitle(
                                blockadeCost,
                                canAffordBlockade,
                                hasBlockadeTarget
                                    ? "Build Blockade"
                                    : "No valid blockade target",
                            )}
                            onSelect={selectBuildAction}
                        />
                    </div>

                    <div class="mt-3 grid grid-cols-2 gap-3">
                        <BuildButton
                            action="floodworks"
                            label="Floodworks"
                            icon="≈"
                            cost={floodworksCost}
                            affordable={canAffordFloodworks}
                            enabled={canBuyFloodworks}
                            variant="water"
                            title={costTitle(
                                floodworksCost,
                                canAffordFloodworks,
                                "Buy 3 flood tokens",
                            )}
                            hint={costTitle(
                                floodworksCost,
                                canAffordFloodworks,
                                "Buy 3 flood tokens",
                            )}
                            onInstant={handleInstantBuild}
                        />

                        <BuildButton
                            action="flood"
                            label="Flood"
                            icon="≈"
                            enabled={canUseFloodToken}
                            selected={selectedBuildAction === "flood"}
                            title={canUseFloodToken
                                ? "Convert one tile to river"
                                : "Need flood token and valid target"}
                            hint={canUseFloodToken
                                ? "Convert one tile to river"
                                : "Need flood token and valid target"}
                            tokenText={`Tokens: ${me?.FloodTokens ?? 0}`}
                            onSelect={selectBuildAction}
                        />
                    </div>

                    <button
                        class="mt-4 w-full cursor-pointer rounded-2xl bg-[#f8efe0]/10 px-5 py-3 font-black text-[#fff7e8] shadow-[0_6px_0_rgba(0,0,0,0.18)] ring-1 ring-[#f8efe0]/20 transition hover:bg-[#f8efe0]/16 active:translate-y-1"
                        type="button"
                        on:click={onPassBuild}
                    >
                        Pass
                    </button>
                </Panel>
            {/if}

            {#if error}
                <Panel>
                    <div
                        class="rounded-2xl bg-[#b94b3f] px-5 py-3 text-sm font-semibold text-white"
                    >
                        {error}
                    </div>
                </Panel>
            {/if}

            <Market {game} {playerId} {role} {onPick} />
        </aside>
    </section>
    <ResourceFlightLayer flights={resourceFlights} />
</AppShell>

<style>
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
