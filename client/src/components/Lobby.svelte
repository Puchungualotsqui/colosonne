<script lang="ts">
    import type { RoomState } from "../lib/types";

    export let room: RoomState;
    export let myPlayerId = 0;
    export let myRole: "player" | "spectator" | "" = "";
    export let loading = false;
    export let error = "";

    export let onReady: (ready: boolean) => void;
    export let onStartGame: () => void;
    export let onKickPlayer: (playerId: number) => void;
    export let onLeaveRoom: () => void;
    export let onCopyRoomCode: () => void;

    $: me = room.players.find((p) => p.playerId === myPlayerId);
    $: isHost = !!me?.isHost;
    $: isReady = !!me?.ready;
    $: canReady = myRole === "player";
    $: enoughPlayers = room.players.length >= 2;
    $: allReady = enoughPlayers && room.players.every((p) => p.ready);
    $: canStart =
        isHost && enoughPlayers && allReady && room.status === "lobby";
</script>

<main
    class="relative min-h-screen overflow-hidden bg-[#17313a] font-sans text-[#f8efe0]"
>
    <div class="pointer-events-none absolute inset-0 bg-[#15323a]">
        <div class="absolute inset-0 bg-board-flat"></div>
        <div class="absolute inset-0 bg-board-texture opacity-[0.16]"></div>
        <div class="absolute inset-x-0 top-0 h-px bg-white/10"></div>
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
                    Game Lobby
                </div>
            </div>
        </div>

        <button
            class="cursor-pointer rounded-xl bg-[#b94b3f] px-4 py-2 text-sm font-bold text-white shadow-sm transition hover:bg-[#c9574a]"
            type="button"
            on:click={onLeaveRoom}
        >
            Leave Room
        </button>
    </header>

    <section
        class="relative z-10 mx-auto grid max-w-7xl gap-6 px-6 pb-12 pt-4 lg:grid-cols-[1fr_360px] lg:px-12"
    >
        <div class="space-y-6">
            <div
                class="rounded-[34px] bg-[#caa66d] p-4 shadow-[0_18px_0_rgba(44,31,21,0.28)] ring-1 ring-black/20"
            >
                <div
                    class="rounded-[26px] border border-[#6b4a2f]/35 bg-[#ead7aa] p-6 text-[#142833] shadow-inner"
                >
                    <div
                        class="flex flex-col gap-5 md:flex-row md:items-center md:justify-between"
                    >
                        <div>
                            <div
                                class="text-sm font-black uppercase tracking-[0.22em] text-[#6b4a2f]"
                            >
                                Room Code
                            </div>

                            <div
                                class="mt-2 inline-flex items-center gap-3 rounded-2xl bg-[#f8efe0] px-5 py-3 shadow-sm ring-1 ring-[#6b4a2f]/15"
                            >
                                <span
                                    class="text-3xl font-black tracking-[0.18em] text-[#142833]"
                                >
                                    {room.roomId}
                                </span>

                                <button
                                    class="cursor-pointer rounded-xl bg-[#f2c36b] px-3 py-2 text-sm font-black text-[#142833] transition hover:bg-[#ffd27c]"
                                    type="button"
                                    on:click={onCopyRoomCode}
                                >
                                    Copy
                                </button>
                            </div>
                        </div>

                        <div class="text-left md:text-right">
                            <div
                                class="text-sm font-black uppercase tracking-[0.22em] text-[#6b4a2f]"
                            >
                                Status
                            </div>

                            <div class="mt-2 text-2xl font-black capitalize">
                                {room.status}
                            </div>

                            <div
                                class="mt-1 text-sm font-semibold text-[#6b4a2f]"
                            >
                                {room.players.length}/{room.settings.maxPlayers}
                                players ·
                                {room.spectators.length}
                                spectators
                            </div>
                        </div>
                    </div>

                    <div
                        class="mt-8 grid gap-4 md:grid-cols-[1fr_auto] md:items-end"
                    >
                        <div>
                            <h1
                                class="text-4xl font-black tracking-tight text-[#142833]"
                            >
                                Waiting for players
                            </h1>

                            <p class="mt-3 max-w-2xl text-[#5c4934]">
                                Share the room code with friends. Players must
                                mark themselves ready before the host can start
                                the match.
                            </p>
                        </div>

                        <div class="flex flex-wrap gap-3 md:justify-end">
                            {#if canReady}
                                <button
                                    class={[
                                        "cursor-pointer rounded-2xl px-6 py-4 text-lg font-black shadow-[0_8px_0_rgba(91,48,28,0.28)] transition hover:-translate-y-0.5 active:translate-y-1 disabled:cursor-not-allowed disabled:opacity-60",
                                        isReady
                                            ? "bg-[#73c4bd] text-[#102b38]"
                                            : "bg-[#f2c36b] text-[#142833]",
                                    ].join(" ")}
                                    type="button"
                                    disabled={loading}
                                    on:click={() => onReady(!isReady)}
                                >
                                    {isReady ? "Ready ✓" : "Ready Up"}
                                </button>
                            {/if}

                            {#if isHost}
                                <button
                                    class="cursor-pointer rounded-2xl bg-[#c96f3d] px-6 py-4 text-lg font-black text-white shadow-[0_8px_0_rgba(91,48,28,0.55)] transition hover:-translate-y-0.5 hover:bg-[#dc7b45] active:translate-y-1 disabled:cursor-not-allowed disabled:opacity-50"
                                    type="button"
                                    disabled={!canStart || loading}
                                    on:click={onStartGame}
                                >
                                    Start Game
                                </button>
                            {/if}
                        </div>
                    </div>

                    {#if !enoughPlayers}
                        <div
                            class="mt-6 rounded-2xl bg-[#f2c36b]/35 p-4 text-sm font-bold text-[#5c4934]"
                        >
                            Need at least 2 players to start.
                        </div>
                    {:else if !allReady}
                        <div
                            class="mt-6 rounded-2xl bg-[#73c4bd]/25 p-4 text-sm font-bold text-[#31545b]"
                        >
                            Waiting for all players to be ready.
                        </div>
                    {:else if !isHost}
                        <div
                            class="mt-6 rounded-2xl bg-[#73c4bd]/25 p-4 text-sm font-bold text-[#31545b]"
                        >
                            All players are ready. Waiting for the host to
                            start.
                        </div>
                    {/if}

                    {#if error}
                        <div
                            class="mt-5 rounded-2xl bg-[#b94b3f] px-5 py-3 font-semibold text-white"
                        >
                            {error}
                        </div>
                    {/if}
                </div>
            </div>

            <div class="grid gap-6 lg:grid-cols-2">
                <section
                    class="rounded-3xl bg-[#23444c] p-5 shadow-md ring-1 ring-[#f8efe0]/10"
                >
                    <div class="flex items-center justify-between">
                        <h2 class="text-xl font-black text-[#fff7e8]">
                            Players
                        </h2>

                        <div
                            class="rounded-xl bg-[#f8efe0]/10 px-3 py-1 text-sm font-bold text-[#9fc9c5]"
                        >
                            {room.players.length}/{room.settings.maxPlayers}
                        </div>
                    </div>

                    <div class="mt-4 space-y-3">
                        {#each room.players as player}
                            <div
                                class="flex items-center justify-between gap-3 rounded-2xl bg-[#f8efe0]/10 p-4 ring-1 ring-[#f8efe0]/10"
                            >
                                <div class="flex min-w-0 items-center gap-3">
                                    <div
                                        class={[
                                            "grid h-11 w-11 place-items-center rounded-2xl text-sm font-black text-white shadow-sm",
                                            player.playerId === 1
                                                ? "bg-[#1d4e89]"
                                                : "bg-[#b94b3f]",
                                        ].join(" ")}
                                    >
                                        P{player.playerId}
                                    </div>

                                    <div class="min-w-0">
                                        <div
                                            class="truncate font-bold text-[#fff7e8]"
                                        >
                                            {player.name}
                                            {#if player.isHost}
                                                <span
                                                    class="ml-2 rounded-lg bg-[#f2c36b] px-2 py-0.5 text-xs font-black text-[#142833]"
                                                >
                                                    Host
                                                </span>
                                            {/if}
                                        </div>

                                        <div class="text-sm text-[#9fc9c5]">
                                            {player.isGuest
                                                ? "Guest"
                                                : "Account"}
                                            ·
                                            {player.ready
                                                ? "Ready"
                                                : "Not ready"}
                                        </div>
                                    </div>
                                </div>

                                <div class="flex items-center gap-2">
                                    <div
                                        class={[
                                            "rounded-xl px-3 py-1 text-sm font-black",
                                            player.ready
                                                ? "bg-[#73c4bd] text-[#102b38]"
                                                : "bg-[#b94b3f]/80 text-white",
                                        ].join(" ")}
                                    >
                                        {player.ready ? "Ready" : "Waiting"}
                                    </div>

                                    {#if isHost && !player.isHost}
                                        <button
                                            class="cursor-pointer rounded-xl bg-[#b94b3f] px-3 py-1 text-sm font-bold text-white transition hover:bg-[#c9574a]"
                                            type="button"
                                            on:click={() =>
                                                onKickPlayer(player.playerId)}
                                        >
                                            Kick
                                        </button>
                                    {/if}
                                </div>
                            </div>
                        {/each}

                        {#if room.players.length < room.settings.maxPlayers}
                            <div
                                class="rounded-2xl border-2 border-dashed border-[#f8efe0]/20 p-4 text-center text-sm font-bold text-[#9fc9c5]"
                            >
                                Empty player slot
                            </div>
                        {/if}
                    </div>
                </section>

                <section
                    class="rounded-3xl bg-[#23444c] p-5 shadow-md ring-1 ring-[#f8efe0]/10"
                >
                    <div class="flex items-center justify-between">
                        <h2 class="text-xl font-black text-[#fff7e8]">
                            Spectators
                        </h2>

                        <div
                            class="rounded-xl bg-[#f8efe0]/10 px-3 py-1 text-sm font-bold text-[#9fc9c5]"
                        >
                            {room.spectators.length}
                        </div>
                    </div>

                    <div class="mt-4 space-y-3">
                        {#if room.spectators.length === 0}
                            <div
                                class="rounded-2xl border-2 border-dashed border-[#f8efe0]/20 p-4 text-center text-sm font-bold text-[#9fc9c5]"
                            >
                                No spectators yet
                            </div>
                        {:else}
                            {#each room.spectators as spectator}
                                <div
                                    class="flex items-center justify-between rounded-2xl bg-[#f8efe0]/10 p-4 ring-1 ring-[#f8efe0]/10"
                                >
                                    <div>
                                        <div class="font-bold text-[#fff7e8]">
                                            {spectator.name}
                                        </div>

                                        <div class="text-sm text-[#9fc9c5]">
                                            {spectator.isGuest
                                                ? "Guest"
                                                : "Account"}
                                        </div>
                                    </div>

                                    <div
                                        class="rounded-xl bg-[#73c4bd]/20 px-3 py-1 text-sm font-bold text-[#9fc9c5]"
                                    >
                                        Watching
                                    </div>
                                </div>
                            {/each}
                        {/if}
                    </div>
                </section>
            </div>
        </div>

        <aside class="space-y-6">
            <section
                class="rounded-3xl bg-[#23444c] p-5 shadow-md ring-1 ring-[#f8efe0]/10"
            >
                <h2 class="text-xl font-black text-[#fff7e8]">
                    Match Settings
                </h2>

                <div class="mt-4 space-y-4">
                    <div
                        class="rounded-2xl bg-[#f8efe0]/10 p-4 ring-1 ring-[#f8efe0]/10"
                    >
                        <div class="text-sm font-bold text-[#9fc9c5]">
                            Max players
                        </div>

                        <div class="mt-1 text-2xl font-black text-[#fff7e8]">
                            {room.settings.maxPlayers}
                        </div>

                        <div class="mt-2 text-sm text-[#b9d5d1]">
                            The current prototype is tuned for two players.
                        </div>
                    </div>

                    <div
                        class="rounded-2xl bg-[#f8efe0]/10 p-4 ring-1 ring-[#f8efe0]/10"
                    >
                        <div class="text-sm font-bold text-[#9fc9c5]">
                            Spectators
                        </div>

                        <div class="mt-1 text-2xl font-black text-[#fff7e8]">
                            {room.settings.spectators ? "Allowed" : "Off"}
                        </div>

                        <div class="mt-2 text-sm text-[#b9d5d1]">
                            Spectators can watch the game but cannot make moves.
                        </div>
                    </div>
                </div>

                {#if isHost}
                    <div
                        class="mt-4 rounded-2xl bg-[#f2c36b]/15 p-4 text-sm font-bold text-[#f8efe0]"
                    >
                        Host settings editing can be added next. For now this
                        lobby locks the MVP to a 2-player match.
                    </div>
                {/if}
            </section>

            <section
                class="rounded-3xl bg-[#23444c] p-5 shadow-md ring-1 ring-[#f8efe0]/10"
            >
                <h2 class="text-xl font-black text-[#fff7e8]">Rules Preview</h2>

                <ul class="mt-4 space-y-3 text-sm font-semibold text-[#b9d5d1]">
                    <li>Draft cards in reverse turn order.</li>
                    <li>Place and build in normal turn order.</li>
                    <li>Influence resolves after all players act.</li>
                    <li>Forests extend outpost influence.</li>
                    <li>Mountains reward connected regions.</li>
                    <li>Cities grow from plains.</li>
                    <li>Rivers block construction, bridges cross them.</li>
                </ul>
            </section>
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
</style>
