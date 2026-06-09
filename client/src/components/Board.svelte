<script lang="ts">
    import Panel from "./ui/Panel.svelte";
    import {
        biomeClass,
        biomeLabel,
        ownerClass,
        ownerLabel,
        structureIcon,
        structureLabel,
    } from "../lib/boardPresentation";
    import {
        canBuildOnTile,
        canReceiveInfluence,
        canUseDraftOnTile,
        controlsTile,
        draftNeedsBoardTarget,
        hasAdjacentControlledTile,
        hexNeighbors,
        isEnemyControlledTile,
        isNeutralTile,
        isUnbridgedRiver,
        tileKey,
        type Coord,
    } from "../lib/rules";
    import {
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
    import BoardTile from "./BoardTile.svelte";

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

    type InfluenceTooltipRow = {
        playerId: number;
        influence: number;
        tempInfluence: number;
        total: number;
        leading: boolean;
    };

    type RenderHex = Coord & {
        key: string;
        left: number;
        top: number;
        tile?: Tile;
        candidate: boolean;
        ghost: boolean;
    };

    type InfluenceLeader = {
        playerId: number;
        value: number;
        tied: boolean;
        total: number;
    };

    type TargetIntent =
        | { kind: "none" }
        | { kind: "placeTile"; handIndex: number }
        | { kind: "draft"; handIndex: number; item: DraftItem }
        | { kind: "build"; action: TargetBuildAction };

    type HexTargetStatus = {
        clickable: boolean;
        dimmed: boolean;
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

    $: debugLog("board.targeting.state", {
        isMyTurn,
        currentPhase: game.CurrentPhase,
        selectedBuildAction,
        selectedHandIndex,
        selectedHandItem,
    });

    function key(x: number, y: number) {
        return tileKey(x, y);
    }

    function isValidTargetForIntent(
        intent: TargetIntent,
        tile: Tile | undefined,
    ) {
        if (!tile) return false;

        switch (intent.kind) {
            case "draft":
                return canUseDraftOnTile(game, playerId, intent.item, tile);

            case "build":
                return canBuildOnTile(game, playerId, intent.action, tile);

            default:
                return false;
        }
    }

    function getHexTargetStatus(hex: RenderHex): HexTargetStatus {
        if (canPlaceTile) {
            return {
                clickable: hex.candidate && !hex.tile && !hex.ghost,
                dimmed: false,
            };
        }

        if (canUseDraft && selectedHandItem && hex.tile) {
            const valid = canUseDraftOnTile(
                game,
                playerId,
                selectedHandItem,
                hex.tile,
            );

            return {
                clickable: valid,
                dimmed: !valid,
            };
        }

        if (canBuild && selectedBuildAction && hex.tile) {
            const valid = canBuildOnTile(
                game,
                playerId,
                selectedBuildAction,
                hex.tile,
            );

            return {
                clickable: valid,
                dimmed: !valid,
            };
        }

        return {
            clickable: false,
            dimmed: false,
        };
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

    function influenceEntries(tile: Tile | undefined) {
        if (!tile) return [];

        return Object.entries(tile.Influence ?? {})
            .map(([playerId, value]) => ({
                playerId: Number(playerId),
                value: Number(value),
            }))
            .filter((entry) => entry.value > 0);
    }

    function influenceLeader(tile: Tile | undefined): InfluenceLeader | null {
        const entries = influenceEntries(tile);

        if (entries.length === 0) return null;

        let bestPlayer = 0;
        let bestValue = 0;
        let tied = false;
        let total = 0;

        for (const entry of entries) {
            total += entry.value;

            if (entry.value > bestValue) {
                bestPlayer = entry.playerId;
                bestValue = entry.value;
                tied = false;
            } else if (entry.value === bestValue) {
                tied = true;
            }
        }

        return {
            playerId: bestPlayer,
            value: bestValue,
            tied,
            total,
        };
    }

    function tileAuraOwner(tile: Tile | undefined) {
        if (!tile) return 0;

        const leader = influenceLeader(tile);

        // If visible influence exists, use it.
        if (leader) {
            if (leader.tied) return 3;
            if (leader.playerId === 1) return 1;
            if (leader.playerId === 2) return 2;
        }

        // Fallback: use ownership.
        // This is what makes the empire outline visible even after Influence is empty
        // or only recalculated temporarily.
        if (tile.HasOwner) {
            if (tile.Owner === 1) return 1;
            if (tile.Owner === 2) return 2;
        }

        return 0;
    }

    function sameAuraGroup(a: Tile | undefined, b: Tile | undefined) {
        const ownerA = tileAuraOwner(a);
        const ownerB = tileAuraOwner(b);

        // Tied influence should not connect into a player empire mass.
        if (ownerA === 3 || ownerB === 3) return false;

        return ownerA > 0 && ownerA === ownerB;
    }

    function auraEdgesForTile(tile: Tile | undefined) {
        if (!tile) return [false, false, false, false, false, false];

        const owner = tileAuraOwner(tile);
        if (owner === 0) return [false, false, false, false, false, false];

        // Neighbor order from hexNeighbors:
        // 0 E, 1 NE, 2 NW, 3 W, 4 SW, 5 SE
        //
        // Edge order in BoardTile:
        // 0 top-right, 1 right, 2 bottom-right, 3 bottom-left, 4 left, 5 top-left
        const neighborToEdge = [1, 0, 5, 4, 3, 2];

        const edges = [false, false, false, false, false, false];
        const neighbors = hexNeighbors(tile.X, tile.Y);

        for (let i = 0; i < neighbors.length; i++) {
            const n = neighbors[i];
            const neighborTile = game.Map.find(
                (t) => t.X === n.x && t.Y === n.y,
            );
            const edgeIndex = neighborToEdge[i];

            // Show an edge only when this side touches non-matching territory,
            // no tile, enemy tile, neutral tile, or tied area.
            edges[edgeIndex] = !sameAuraGroup(tile, neighborTile);
        }

        return edges;
    }

    function blockadeTargetMessage(tile: Tile) {
        if (tile.Biome === Biome.River) return "Cannot blockade river";
        if (tile.HasBlockade) return "Already blockaded";
        if (controlsTile(tile, playerId))
            return "Cannot blockade your own tile";

        if (!hasAdjacentControlledTile(game, tile.X, tile.Y, playerId)) {
            return "Needs adjacent controlled tile";
        }

        if (isEnemyControlledTile(tile, playerId)) {
            return tile.Structure !== Structure.None
                ? "Blockade enemy structure"
                : "Blockade enemy tile";
        }

        if (isNeutralTile(tile)) {
            return "Blockade neutral tile";
        }

        return "Invalid blockade target";
    }

    function settlementTargetMessage(tile: Tile) {
        if (tile.Biome === Biome.River) {
            return "Settlement cannot be built on river";
        }

        if (tile.Structure !== Structure.None) {
            return "Tile already has a structure";
        }

        if (!controlsTile(tile, playerId)) {
            return "Settlement requires friendly territory";
        }

        return "Build Settlement";
    }

    function tileTooltip(tile: Tile | undefined, candidate: boolean) {
        if (candidate) return "Place tile";
        if (!tile) return "";

        if (canUseDraft && selectedHandItem) {
            if (canUseDraftOnTile(game, playerId, selectedHandItem, tile)) {
                if (
                    selectedHandItem.Kind === DraftKind.Structure &&
                    selectedHandItem.Structure === Structure.Watchtower
                ) {
                    return tile.HasOwner
                        ? "Build Watchtower on your tile"
                        : "Build Watchtower on neutral tile";
                }

                if (
                    selectedHandItem.Kind === DraftKind.Structure &&
                    selectedHandItem.Structure === Structure.Bridge
                ) {
                    return "Build Bridge";
                }

                return "Use card here";
            }

            if (
                selectedHandItem.Kind === DraftKind.Structure &&
                selectedHandItem.Structure === Structure.Watchtower
            ) {
                if (tile.Biome === Biome.River)
                    return "Watchtower cannot be built on river";
                if (tile.Structure !== Structure.None)
                    return "Tile already has a structure";
                if (isEnemyControlledTile(tile, playerId)) {
                    return "Cannot build Watchtower on enemy tile";
                }
            }
        }

        if (canBuild && selectedBuildAction) {
            if (canBuildOnTile(game, playerId, selectedBuildAction, tile)) {
                switch (selectedBuildAction) {
                    case "outpost":
                        return "Build Outpost";

                    case "settlement":
                        return "Build Settlement";

                    case "city":
                        return "Upgrade Outpost to City";

                    case "blockade":
                        return blockadeTargetMessage(tile);

                    case "flood":
                        return "Convert to River";
                }
            }

            switch (selectedBuildAction) {
                case "outpost":
                    if (tile.Biome === Biome.River) {
                        return "Outpost cannot be built on River";
                    }

                    if (tile.Structure !== Structure.None) {
                        return "Tile already has a structure";
                    }

                    if (tile.HasBlockade) {
                        return "Cannot build on blockaded tile";
                    }

                    if (isEnemyControlledTile(tile, playerId)) {
                        return "Cannot build Outpost on enemy tile";
                    }

                    return "Invalid Outpost target";

                case "settlement":
                    return settlementTargetMessage(tile);

                case "city":
                    if (tile.Structure !== Structure.Outpost) {
                        return "City requires your own Outpost";
                    }

                    if (tile.StructureOwner !== playerId) {
                        return "Requires your own Outpost";
                    }

                    if (tile.HasBlockade) {
                        return "Cannot upgrade blockaded tile";
                    }

                    if (!controlsTile(tile, playerId)) {
                        return "Outpost must be active and controlled by you";
                    }

                    return "Upgrade Outpost to City";

                case "blockade":
                    return blockadeTargetMessage(tile);

                case "flood":
                    if (tile.Biome === Biome.River) {
                        return "Already river";
                    }

                    if (tile.Structure !== Structure.None) {
                        return "Cannot flood a structure";
                    }

                    return "Invalid Flood target";
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

    function handleHexSelect(hex: RenderHex) {
        const status = getHexTargetStatus(hex);
        const intent = getTargetIntent();

        debugLog("board.hex.select", {
            x: hex.x,
            y: hex.y,
            hasTile: !!hex.tile,
            candidate: hex.candidate,
            ghost: hex.ghost,
            role,
            playerId,
            currentPlayer: game.CurrentPlayer,
            currentPhase: game.CurrentPhase,
            isMyTurn,
            canPlaceTile,
            canUseDraft,
            canBuild,
            selectedBuildAction,
            selectedHandIndex,
            selectedHandItem,
            freshIntent: intent,
            clickable: status.clickable,
            dimmed: status.dimmed,
            tile: hex.tile,
        });

        if (!status.clickable) return;

        switch (intent.kind) {
            case "placeTile":
                if (!hex.candidate) return;
                onPlaceTile(intent.handIndex, hex.x, hex.y);
                return;

            case "draft":
                if (!hex.tile) return;
                onUseDraft(intent.handIndex, hex.tile.X, hex.tile.Y);
                return;

            case "build":
                if (!hex.tile) return;
                onBuild(intent.action, hex.tile.X, hex.tile.Y);
                return;

            case "none":
                return;
        }
    }

    function getTargetIntent(): TargetIntent {
        if (
            isMyTurn &&
            game.CurrentPhase === GamePhase.Place &&
            selectedHandIndex >= 0 &&
            selectedHandItem?.Kind === DraftKind.Tile
        ) {
            return {
                kind: "placeTile",
                handIndex: selectedHandIndex,
            };
        }

        if (
            isMyTurn &&
            game.CurrentPhase === GamePhase.Place &&
            selectedHandIndex >= 0 &&
            selectedHandItem !== null &&
            selectedHandItem.Kind !== DraftKind.Tile &&
            draftNeedsBoardTarget(selectedHandItem)
        ) {
            return {
                kind: "draft",
                handIndex: selectedHandIndex,
                item: selectedHandItem,
            };
        }

        if (
            isMyTurn &&
            game.CurrentPhase === GamePhase.Build &&
            selectedBuildAction !== null
        ) {
            return {
                kind: "build",
                action: selectedBuildAction,
            };
        }

        return { kind: "none" };
    }

    let hoveredStructureKey = "";
    let hoveredStructureInfluence = new Set<string>();
    let hoveredStructureOwner = 0;

    function tileAt(x: number, y: number) {
        return game.Map.find((tile) => tile.X === x && tile.Y === y);
    }

    function hexRingRadius2(x: number, y: number): Coord[] {
        const results: Coord[] = [];

        for (let dx = -2; dx <= 2; dx++) {
            for (let dy = -2; dy <= 2; dy++) {
                const dz = -dx - dy;

                if (
                    (Math.abs(dx) === 2 ||
                        Math.abs(dy) === 2 ||
                        Math.abs(dz) === 2) &&
                    Math.abs(dx) <= 2 &&
                    Math.abs(dy) <= 2 &&
                    Math.abs(dz) <= 2
                ) {
                    results.push({ x: x + dx, y: y + dy });
                }
            }
        }

        return results;
    }

    function canInfluenceDistance2Preview(
        fromX: number,
        fromY: number,
        toX: number,
        toY: number,
    ) {
        for (const a of hexNeighbors(fromX, fromY)) {
            for (const b of hexNeighbors(toX, toY)) {
                if (a.x !== b.x || a.y !== b.y) continue;

                const middle = tileAt(a.x, a.y);
                if (!middle) continue;

                if (!isUnbridgedRiver(middle)) {
                    return true;
                }
            }
        }

        return false;
    }

    function structureInfluenceCoords(tile: Tile | undefined) {
        if (!tile) return [];

        const coords: Coord[] = [];

        function addCoord(x: number, y: number) {
            const target = tileAt(x, y);
            if (!canReceiveInfluence(target)) return;

            coords.push({ x, y });
        }

        switch (tile.Structure) {
            case Structure.Settlement:
                addCoord(tile.X, tile.Y);
                break;

            case Structure.Outpost:
                addCoord(tile.X, tile.Y);

                for (const n of hexNeighbors(tile.X, tile.Y)) {
                    addCoord(n.x, n.y);
                }

                break;

            case Structure.City:
                addCoord(tile.X, tile.Y);
                break;

            case Structure.Watchtower:
                addCoord(tile.X, tile.Y);

                for (const n of hexNeighbors(tile.X, tile.Y)) {
                    addCoord(n.x, n.y);
                }

                for (const n of hexRingRadius2(tile.X, tile.Y)) {
                    if (
                        canInfluenceDistance2Preview(tile.X, tile.Y, n.x, n.y)
                    ) {
                        addCoord(n.x, n.y);
                    }
                }

                break;

            case Structure.Bridge:
                addCoord(tile.X, tile.Y);

                for (const n of hexNeighbors(tile.X, tile.Y)) {
                    addCoord(n.x, n.y);
                }

                break;
        }

        return coords;
    }

    function handleStructureHover(tile: Tile | undefined) {
        if (!tile || tile.Structure === Structure.None) {
            hoveredStructureKey = "";
            hoveredStructureInfluence = new Set();
            hoveredStructureOwner = 0;
            return;
        }

        hoveredStructureKey = key(tile.X, tile.Y);
        hoveredStructureOwner = tile.StructureOwner;

        hoveredStructureInfluence = new Set(
            structureInfluenceCoords(tile).map((coord) =>
                key(coord.x, coord.y),
            ),
        );
    }

    function clearStructureHover() {
        hoveredStructureKey = "";
        hoveredStructureInfluence = new Set();
        hoveredStructureOwner = 0;
    }

    function isStructurePreviewed(hex: RenderHex) {
        if (!hex.tile) return false;
        if (!hoveredStructureKey) return false;

        return hoveredStructureInfluence.has(key(hex.tile.X, hex.tile.Y));
    }

    function tileInfluenceRows(tile: Tile | undefined): InfluenceTooltipRow[] {
        if (!tile) return [];

        const rows = game.Players.map((player) => {
            const influence = Number(tile.Influence?.[player.Id] ?? 0);
            const tempInfluence = Number(tile.TempInfluence?.[player.Id] ?? 0);

            return {
                playerId: player.Id,
                influence,
                tempInfluence,
                total: influence + tempInfluence,
                leading: false,
            };
        }).filter((row) => row.total > 0);

        if (rows.length === 0) return [];

        const best = Math.max(...rows.map((row) => row.total));

        return rows
            .map((row) => ({
                ...row,
                leading: row.total === best && best > 0,
            }))
            .sort((a, b) => b.total - a.total);
    }

    function tileTooltipTitle(tile: Tile | undefined, candidate: boolean) {
        if (candidate) return "Place tile";
        if (!tile) return "";

        const biome = biomeLabel(tile.Biome);
        const owner = tile.HasOwner ? `P${tile.Owner}` : "Open";
        const structure =
            tile.Structure !== Structure.None
                ? structureLabel(tile.Structure)
                : "No structure";

        return `${biome} · ${owner}`;
    }

    function tileTooltipSubtitle(tile: Tile | undefined, candidate: boolean) {
        if (candidate) return "Empty placement candidate";
        if (!tile) return "";

        const parts: string[] = [];

        if (tile.Structure !== Structure.None) {
            parts.push(structureLabel(tile.Structure));
        }

        if (tile.HasBlockade) {
            parts.push(`Blockaded by P${tile.BlockadeOwner}`);
        }

        if (tile.Biome === Biome.River && tile.Structure !== Structure.Bridge) {
            parts.push("Unbridged river");
        }

        return parts.join(" · ");
    }
</script>

<Panel
    variant="board"
    padding="sm"
    className="min-h-[calc(100vh-132px)] min-w-0 overflow-hidden"
>
    <Panel variant="innerBoard" padding="md">
        <div class="mb-4 flex items-center justify-between gap-3">
            <h2 class="text-xl font-black text-[#142833]">Board</h2>

            <div
                class="rounded-xl bg-[#f8efe0] px-3 py-1 text-sm font-black text-[#142833]"
            >
                {game.Map.length} tiles
            </div>
        </div>

        <div
            class="h-[calc(100vh-230px)] max-w-full overflow-auto rounded-2xl bg-[#d9c291] p-4 shadow-inner"
        >
            <div
                class="relative"
                style={`width: ${boardSize.width}px; height: ${boardSize.height}px;`}
            >
                {#each renderHexes as hex}
                    {@const targetStatus = getHexTargetStatus(hex)}

                    <BoardTile
                        left={hex.left}
                        top={hex.top}
                        width={HEX_W}
                        height={HEX_H}
                        tile={hex.tile}
                        candidate={hex.candidate}
                        ghost={hex.ghost}
                        clickable={targetStatus.clickable}
                        dimmed={targetStatus.dimmed}
                        biomeClass={biomeClass(hex.tile, hex.candidate)}
                        ownerClass={hex.tile ? ownerClass(hex.tile) : ""}
                        ownerLabel={hex.tile ? ownerLabel(hex.tile) : ""}
                        structureLabel={hex.tile
                            ? structureLabel(hex.tile.Structure)
                            : ""}
                        structureIcon={hex.tile
                            ? structureIcon(hex.tile.Structure)
                            : ""}
                        tooltip={tileTooltip(hex.tile, hex.candidate)}
                        tooltipTitle={tileTooltipTitle(hex.tile, hex.candidate)}
                        tooltipSubtitle={tileTooltipSubtitle(
                            hex.tile,
                            hex.candidate,
                        )}
                        influenceRows={tileInfluenceRows(hex.tile)}
                        auraOwner={tileAuraOwner(hex.tile)}
                        auraEdges={auraEdgesForTile(hex.tile)}
                        influencePreviewed={isStructurePreviewed(hex)}
                        influencePreviewOwner={hoveredStructureOwner}
                        onStructureHover={() => handleStructureHover(hex.tile)}
                        onStructureLeave={clearStructureHover}
                        on:select={() => handleHexSelect(hex)}
                    />
                {/each}
            </div>
        </div>
    </Panel>
</Panel>
