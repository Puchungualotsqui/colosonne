<script lang="ts">
    import {
        Action,
        Biome,
        DraftKind,
        GamePhase,
        Structure,
        type DraftItem,
        type GameState,
        type TargetBuildAction,
        type Tile,
    } from "../lib/types";
    import { debugLog } from "../lib/debug";

    export let game: GameState;
    export let playerId = 0;
    export let role: "player" | "spectator" | "" = "";
    export let selectedBuildAction: TargetBuildAction | null = null;

    export let selectedHandIndex = -1;
    export let selectedHandItem: DraftItem | null = null;

    export let onPlaceTile: (handIndex: number, x: number, y: number) => void;
    export let onUseDraft: (
        handIndex: number,
        x: number,
        y: number,
        targetPlayerId?: number,
    ) => void;
    export let onBuild: (
        action: TargetBuildAction,
        x: number,
        y: number,
    ) => void;

    type Coord = {
        x: number;
        y: number;
    };

    type RenderHex = Coord & {
        key: string;
        left: number;
        top: number;
        tile?: Tile;
        candidate: boolean;
        ghost: boolean;
    };

    const HEX_W = 112;
    const HEX_H = 98;
    const STEP_X = 93;
    const STEP_Y = 74;
    const PAD = 90;

    $: isMyTurn = role === "player" && game.CurrentPlayer === playerId;

    $: canPlaceTile =
        isMyTurn &&
        game.CurrentPhase === GamePhase.Place &&
        selectedHandIndex >= 0 &&
        selectedHandItem?.Kind === DraftKind.Tile;

    $: canUseDraft =
        isMyTurn &&
        game.CurrentPhase === GamePhase.Place &&
        selectedHandIndex >= 0 &&
        selectedHandItem !== null &&
        selectedHandItem.Kind !== DraftKind.Tile &&
        draftNeedsBoardTarget(selectedHandItem);

    $: canBuild =
        isMyTurn &&
        game.CurrentPhase === GamePhase.Build &&
        selectedBuildAction !== null;

    $: placementShell = placementCandidates(game.Map);
    $: renderHexes = buildRenderHexes(game.Map, placementShell, canPlaceTile);
    $: boardSize = calculateBoardSize(renderHexes);

    function key(x: number, y: number) {
        return `${x},${y}`;
    }

    function hexNeighbors(x: number, y: number): Coord[] {
        return [
            { x: x + 1, y },
            { x: x + 1, y: y - 1 },
            { x, y: y - 1 },
            { x: x - 1, y },
            { x: x - 1, y: y + 1 },
            { x, y: y + 1 },
        ];
    }

    function rawPosition(x: number, y: number) {
        return {
            left: (x + y / 2) * STEP_X,
            top: y * STEP_Y,
        };
    }

    function placementCandidates(map: Tile[]): Coord[] {
        const occupied = new Set(map.map((t) => key(t.X, t.Y)));
        const out = new Map<string, Coord>();

        for (const tile of map) {
            for (const n of hexNeighbors(tile.X, tile.Y)) {
                const k = key(n.x, n.y);
                if (!occupied.has(k)) {
                    out.set(k, n);
                }
            }
        }

        return [...out.values()];
    }

    function buildRenderHexes(
        map: Tile[],
        candidateCoords: Coord[],
        showCandidates: boolean,
    ): RenderHex[] {
        const tileByKey = new Map(map.map((t) => [key(t.X, t.Y), t]));

        const all = [
            ...map.map((tile) => ({
                x: tile.X,
                y: tile.Y,
                tile,
                candidate: false,
                ghost: false,
            })),
            ...candidateCoords.map((c) => ({
                x: c.x,
                y: c.y,
                tile: undefined,
                candidate: showCandidates,
                ghost: !showCandidates,
            })),
        ];

        const raw = all.map((item) => {
            const pos = rawPosition(item.x, item.y);
            return {
                ...item,
                key: key(item.x, item.y),
                left: pos.left,
                top: pos.top,
            };
        });

        const minLeft = Math.min(...raw.map((h) => h.left));
        const minTop = Math.min(...raw.map((h) => h.top));

        return raw.map((h) => ({
            ...h,
            left: h.left - minLeft + PAD,
            top: h.top - minTop + PAD,
            tile: tileByKey.get(h.key),
        }));
    }

    function calculateBoardSize(hexes: RenderHex[]) {
        if (hexes.length === 0) {
            return {
                width: 600,
                height: 440,
            };
        }

        const maxLeft = Math.max(...hexes.map((h) => h.left));
        const maxTop = Math.max(...hexes.map((h) => h.top));

        return {
            width: Math.max(620, maxLeft + HEX_W + PAD),
            height: Math.max(440, maxTop + HEX_H + PAD),
        };
    }

    function biomeClass(tile: Tile | undefined, candidate: boolean) {
        if (candidate) {
            return "border-[#d1a45f] bg-[#ead7aa]/35 text-[#6b4a2f]";
        }

        switch (tile?.Biome) {
            case Biome.Forest:
                return "border-[#2f6546] bg-[#5b9368] text-[#17313a]";
            case Biome.Mountain:
                return "border-[#656b73] bg-[#a8adb2] text-[#142833]";
            case Biome.Plain:
                return "border-[#9b7034] bg-[#d9b56a] text-[#142833]";
            case Biome.River:
                return "border-[#327b8d] bg-[#6eb8c5] text-[#102b38]";
            case Biome.Ruins:
                return "border-[#6d4c9b] bg-[#9b79c9] text-[#142833]";
            default:
                return "border-[#6b4a2f] bg-[#ead7aa] text-[#142833]";
        }
    }

    function biomeLabel(biome: Biome | undefined) {
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
                return "";
        }
    }

    function structureLabel(structure: Structure) {
        switch (structure) {
            case Structure.Outpost:
                return "Outpost";
            case Structure.City:
                return "City";
            case Structure.Settlement:
                return "Settlement";
            case Structure.Bridge:
                return "Bridge";
            case Structure.Watchtower:
                return "Watchtower";
            default:
                return "";
        }
    }

    function structureIcon(structure: Structure) {
        switch (structure) {
            case Structure.Outpost:
                return "⌂";
            case Structure.City:
                return "▦";
            case Structure.Settlement:
                return "◈";
            case Structure.Bridge:
                return "⌒";
            case Structure.Watchtower:
                return "♜";
            default:
                return "";
        }
    }

    function ownerClass(tile: Tile) {
        if (!tile.HasOwner) return "bg-[#f8efe0]/80 text-[#5c4934]";

        if (tile.Owner === 1) {
            return "bg-[#1d4e89] text-white";
        }

        if (tile.Owner === 2) {
            return "bg-[#b94b3f] text-white";
        }

        return "bg-[#f8efe0]/80 text-[#5c4934]";
    }

    function ownerLabel(tile: Tile) {
        if (!tile.HasOwner) return "Open";
        return `P${tile.Owner}`;
    }

    function controlsTile(
        tile: { HasOwner: boolean; Owner: number } | undefined,
    ) {
        return !!tile && tile.HasOwner && tile.Owner === playerId;
    }

    function hasAdjacentControlledTile(x: number, y: number) {
        return hexNeighbors(x, y).some((n) => {
            const tile = game.Map.find((t) => t.X === n.x && t.Y === n.y);
            return controlsTile(tile);
        });
    }

    function draftNeedsBoardTarget(item: DraftItem) {
        if (item.Kind === DraftKind.Structure) return true;

        if (item.Kind === DraftKind.Action) {
            return (
                item.Action === Action.Harvest ||
                item.Action === Action.Reinforce
            );
        }

        return false;
    }

    function isValidDraftTarget(
        item: DraftItem | null,
        tile: Tile | undefined,
    ) {
        if (!item || !tile) return false;

        if (item.Kind === DraftKind.Structure) {
            if (tile.Structure !== Structure.None) return false;

            switch (item.Structure) {
                case Structure.Bridge:
                    return (
                        tile.Biome === Biome.River &&
                        hasAdjacentControlledTile(tile.X, tile.Y)
                    );

                case Structure.Watchtower:
                    return tile.Biome !== Biome.River && controlsTile(tile);

                case Structure.Outpost:
                case Structure.City:
                case Structure.Settlement:
                    return false;

                default:
                    return false;
            }
        }

        if (item.Kind === DraftKind.Action) {
            if (item.Action === Action.Harvest) {
                return controlsTile(tile) && tile.Biome !== Biome.River;
            }

            if (item.Action === Action.Reinforce) {
                return true;
            }
        }

        return false;
    }

    function isValidBuildTarget(tile: Tile | undefined) {
        if (!canBuild || !selectedBuildAction || !tile) return false;

        switch (selectedBuildAction) {
            case "outpost":
                return (
                    tile.Biome !== Biome.River &&
                    tile.Structure === Structure.None
                );

            case "settlement":
                return (
                    tile.Biome !== Biome.River &&
                    tile.Structure === Structure.None
                );

            case "city":
                return (
                    tile.Structure === Structure.Outpost &&
                    tile.StructureOwner === playerId
                );

            case "blockade":
                return (
                    tile.Biome !== Biome.River &&
                    tile.Structure === Structure.None &&
                    !tile.HasBlockade
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

    function tileTooltip(tile: Tile | undefined, candidate: boolean) {
        if (candidate) return "Place tile";
        if (!tile) return "";

        if (canUseDraft && selectedHandItem) {
            if (isValidDraftTarget(selectedHandItem, tile)) {
                return "Use card here";
            }
        }

        if (canBuild && selectedBuildAction) {
            if (isValidBuildTarget(tile)) {
                switch (selectedBuildAction) {
                    case "outpost":
                        return "Build Outpost";
                    case "settlement":
                        return "Build Settlement";
                    case "city":
                        return "Upgrade to City";
                    case "blockade":
                        return "Build Blockade";
                    case "flood":
                        return "Convert to River";
                }
            }

            if (selectedBuildAction === "city") {
                return "Requires your outpost";
            }

            if (selectedBuildAction === "flood") {
                if (tile.Biome === Biome.River) return "Already river";
                if (tile.Structure !== Structure.None)
                    return "Cannot flood a structure";
            }

            if (tile.Biome === Biome.River) {
                return tile.Structure === Structure.Bridge
                    ? "Bridged river"
                    : "River blocks normal control";
            }

            if (tile.Structure !== Structure.None) {
                return "Tile already has a structure";
            }
        }

        const owner = tile.HasOwner ? `P${tile.Owner}` : "Open";
        const structure =
            tile.Structure !== Structure.None
                ? structureLabel(tile.Structure)
                : "No structure";

        if (tile.Biome === Biome.River && tile.Structure !== Structure.Bridge) {
            return `River · Not controllable · ${structure}`;
        }

        if (tile.Biome === Biome.Ruins) {
            return `Ruins · ${owner} · ${structure} · Relic`;
        }

        return `${biomeLabel(tile.Biome)} · ${owner} · ${structure}`;
    }

    function handleHexPointerDown(hex: RenderHex) {
        debugLog("board.hex.pointerdown", {
            x: hex.x,
            y: hex.y,
            isClickable: isClickable(hex),
            canBuild,
            selectedBuildAction,
            selectedHandIndex,
            selectedHandItem,
            tile: hex.tile,
        });

        if (!isClickable(hex)) return;
        handleHexClick(hex);
    }

    function handleHexClick(hex: RenderHex) {
        debugLog("board.hex.click", {
            x: hex.x,
            y: hex.y,
            hasTile: !!hex.tile,
            candidate: hex.candidate,
            role,
            playerId,
            currentPlayer: game.CurrentPlayer,
            currentPhase: game.CurrentPhase,
            isMyTurn,
            selectedBuildAction,
            selectedHandIndex,
            selectedHandItem,
            canBuild,
            canPlaceTile,
            canUseDraft,
            isClickable: isClickable(hex),
            isValidBuildTarget: hex.tile ? isValidBuildTarget(hex.tile) : false,
            isValidDraftTarget: hex.tile
                ? isValidDraftTarget(selectedHandItem, hex.tile)
                : false,
            tile: hex.tile,
        });

        if (hex.candidate && canPlaceTile) {
            onPlaceTile(selectedHandIndex, hex.x, hex.y);
            return;
        }

        if (!hex.tile) return;

        if (canUseDraft && isValidDraftTarget(selectedHandItem, hex.tile)) {
            onUseDraft(selectedHandIndex, hex.tile.X, hex.tile.Y);
            return;
        }

        if (canBuild && selectedBuildAction && isValidBuildTarget(hex.tile)) {
            onBuild(selectedBuildAction, hex.tile.X, hex.tile.Y);
        }
    }

    function isClickable(hex: RenderHex) {
        if (hex.candidate && canPlaceTile) return true;

        if (
            hex.tile &&
            canUseDraft &&
            isValidDraftTarget(selectedHandItem, hex.tile)
        ) {
            return true;
        }

        if (hex.tile && canBuild && isValidBuildTarget(hex.tile)) {
            return true;
        }

        return false;
    }
</script>

<section
    class="min-w-0 overflow-hidden rounded-[34px] bg-[#caa66d] p-4 shadow-[0_18px_0_rgba(44,31,21,0.28)] ring-1 ring-black/20"
>
    <div
        class="rounded-[26px] border border-[#6b4a2f]/35 bg-[#ead7aa] p-4 shadow-inner"
    >
        <div class="mb-4 flex items-center justify-between gap-3">
            <h2 class="text-xl font-black text-[#142833]">Board</h2>

            <div
                class="rounded-xl bg-[#f8efe0] px-3 py-1 text-sm font-black text-[#142833]"
            >
                {game.Map.length} tiles
            </div>
        </div>

        <div
            class="max-h-[calc(100vh-220px)] max-w-full overflow-auto rounded-2xl bg-[#d9c291] p-4 shadow-inner"
        >
            <div
                class="relative"
                style={`width: ${boardSize.width}px; height: ${boardSize.height}px;`}
            >
                {#each renderHexes as hex}
                    <button
                        class={[
                            "group clip-hex absolute flex items-center justify-center border-[2px] shadow-[0_7px_0_rgba(74,48,31,0.22)] transition",
                            biomeClass(hex.tile, hex.candidate),
                            isClickable(hex)
                                ? "cursor-pointer ring-4 ring-[#f2c36b] ring-offset-2 ring-offset-[#d9c291] hover:-translate-y-1 hover:brightness-110"
                                : "cursor-default",
                            hex.candidate ? "border-dashed opacity-85" : "",
                            hex.ghost ? "pointer-events-none opacity-0" : "",
                            (canBuild || canUseDraft) &&
                            hex.tile &&
                            !isClickable(hex)
                                ? "opacity-55"
                                : "",
                        ].join(" ")}
                        style={`left: ${hex.left}px; top: ${hex.top}px; width: ${HEX_W}px; height: ${HEX_H}px;`}
                        type="button"
                        disabled={!isClickable(hex)}
                        aria-label={tileTooltip(hex.tile, hex.candidate)}
                        on:pointerdown|preventDefault={() =>
                            handleHexPointerDown(hex)}
                    >
                        <div
                            class="pointer-events-none absolute inset-[5px] clip-hex border border-white/25"
                        ></div>

                        {#if hex.candidate}
                            <div
                                class="relative z-10 grid h-10 w-10 place-items-center rounded-full bg-[#f8efe0]/70 text-2xl font-black text-[#6b4a2f]"
                            >
                                +
                            </div>
                        {:else if hex.tile}
                            {#if hex.tile.Structure !== Structure.None}
                                <div
                                    class="relative z-10 grid h-12 w-12 place-items-center rounded-2xl bg-[#f8efe0]/70 text-2xl font-black text-[#142833] shadow-sm"
                                    title={structureLabel(hex.tile.Structure)}
                                >
                                    {structureIcon(hex.tile.Structure)}
                                </div>
                            {/if}

                            {#if hex.tile.HasBlockade}
                                <div
                                    class={[
                                        "absolute right-3 top-3 z-20 grid h-6 w-6 place-items-center rounded-full text-xs font-black shadow-sm",
                                        hex.tile.BlockadeOwner === 1
                                            ? "bg-[#1d4e89] text-white"
                                            : hex.tile.BlockadeOwner === 2
                                              ? "bg-[#b94b3f] text-white"
                                              : "bg-[#142833] text-white",
                                    ].join(" ")}
                                    title={`Blockade P${hex.tile.BlockadeOwner}`}
                                >
                                    ✕
                                </div>
                            {/if}

                            {#if hex.tile.HasOwner}
                                <div
                                    class={[
                                        "absolute bottom-2 left-1/2 z-20 h-5 min-w-8 -translate-x-1/2 rounded-full px-2 text-[10px] font-black leading-5 shadow-sm",
                                        ownerClass(hex.tile),
                                    ].join(" ")}
                                >
                                    {ownerLabel(hex.tile)}
                                </div>
                            {/if}
                        {/if}

                        <div
                            class="pointer-events-none absolute bottom-[calc(100%+8px)] left-1/2 z-50 hidden w-max max-w-[220px] -translate-x-1/2 rounded-xl bg-[#142833] px-3 py-2 text-xs font-bold text-[#f8efe0] shadow-xl ring-1 ring-white/10 group-hover:block"
                        >
                            {tileTooltip(hex.tile, hex.candidate)}
                        </div>
                    </button>
                {/each}
            </div>
        </div>
    </div>
</section>

<style>
    .clip-hex {
        clip-path: polygon(
            50% 0%,
            93.3% 25%,
            93.3% 75%,
            50% 100%,
            6.7% 75%,
            6.7% 25%
        );
    }
</style>
