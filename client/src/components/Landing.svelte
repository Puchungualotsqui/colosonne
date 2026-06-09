<script lang="ts" context="module">
    import type { StructureGlyphKind } from "./StructureGlyph.svelte";
    import StructureGlyph from "./StructureGlyph.svelte";
    export type LandingUser = {
        authenticated: boolean;
        isGuest: boolean;
        displayName: string;
        avatarUrl?: string;
        karma: number;
    } | null;
</script>

<script lang="ts">
    import AppShell from "./ui/AppShell.svelte";
    import AppHeader from "./ui/AppHeader.svelte";
    import { ui } from "../lib/uiClasses";

    export let user: LandingUser = null;
    export let loading = false;
    export let error = "";

    export let onCreateRoom: () => void;
    export let onJoinRoom: (roomId: string) => void;
    export let onSpectateRoom: (roomId: string) => void;
    export let onLogin: () => void;
    export let onSignUp: () => void;

    let joinOpen = false;
    let roomCode = "";

    type PreviewTile = {
        b: "mountain" | "plain" | "forest" | "river";
        x: number;
        y: number;
        structure?: StructureGlyphKind;
        owner?: "blue" | "red";
    };

    const previewTiles: PreviewTile[] = [
        { b: "mountain", x: 82, y: 0, structure: "watchtower", owner: "blue" },
        { b: "plain", x: 200, y: 0, structure: "city", owner: "red" },
        { b: "forest", x: 318, y: 0 },

        { b: "plain", x: 23, y: 88, structure: "outpost", owner: "blue" },
        { b: "forest", x: 141, y: 88 },
        { b: "river", x: 259, y: 88, structure: "bridge", owner: "red" },
        { b: "mountain", x: 377, y: 88 },

        { b: "forest", x: 82, y: 176 },
        { b: "plain", x: 200, y: 176, structure: "road", owner: "blue" },
        { b: "mountain", x: 318, y: 176, structure: "outpost", owner: "red" },

        { b: "river", x: 23, y: 264 },
        { b: "plain", x: 141, y: 264, structure: "city", owner: "blue" },
        { b: "forest", x: 259, y: 264, structure: "watchtower", owner: "red" },
        { b: "plain", x: 377, y: 264, structure: "road", owner: "red" },
    ];

    function join() {
        const cleaned = roomCode.trim();
        if (!cleaned) return;
        onJoinRoom(cleaned);
    }

    function spectate() {
        const cleaned = roomCode.trim();
        if (!cleaned) return;
        onSpectateRoom(cleaned);
    }
</script>

<AppShell>
    <AppHeader subtitle="Influence Strategy">
        <svelte:fragment slot="actions">
            {#if user && user.authenticated && !user.isGuest}
                <div
                    class="flex items-center gap-3 rounded-2xl bg-[#f8efe0]/10 px-4 py-2 shadow-sm ring-1 ring-[#f8efe0]/15 backdrop-blur"
                >
                    {#if user.avatarUrl}
                        <img
                            class="h-9 w-9 rounded-full object-cover ring-2 ring-[#f2c36b]/70"
                            src={user.avatarUrl}
                            alt={user.displayName}
                        />
                    {:else}
                        <div
                            class="grid h-9 w-9 place-items-center rounded-full bg-[#f2c36b] text-sm font-bold text-[#142833]"
                        >
                            {user.displayName.slice(0, 1).toUpperCase()}
                        </div>
                    {/if}

                    <div class="text-left">
                        <div class="text-sm font-semibold text-[#fff7e8]">
                            {user.displayName}
                        </div>
                        <div class="text-xs text-[#9fc9c5]">
                            Karma {user.karma}
                        </div>
                    </div>
                </div>
            {:else}
                <button
                    class={ui.button.ghost}
                    type="button"
                    on:click={onLogin}
                >
                    Log in
                </button>

                <button
                    class={ui.button.cream}
                    type="button"
                    on:click={onSignUp}
                >
                    Sign up
                </button>
            {/if}
        </svelte:fragment>
    </AppHeader>

    <section
        class="relative z-10 mx-auto grid min-h-[calc(100vh-84px)] max-w-7xl grid-cols-1 items-center gap-10 px-6 pb-12 lg:grid-cols-[1.05fr_0.95fr] lg:px-12"
    >
        <div class="flex justify-center lg:justify-start">
            <div class="relative h-[520px] w-[700px] max-w-full">
                <div
                    class="absolute inset-x-3 top-4 bottom-4 rounded-[34px] bg-[#caa66d] shadow-[0_18px_0_rgba(44,31,21,0.28)] ring-1 ring-black/20"
                ></div>

                <div
                    class="absolute inset-x-8 top-9 bottom-9 rounded-[26px] border border-[#6b4a2f]/35 bg-[#ead7aa] shadow-inner"
                ></div>

                <div
                    class="absolute left-1/2 top-1/2 h-[430px] w-[560px] -translate-x-1/2 -translate-y-1/2"
                >
                    {#each previewTiles as tile}
                        <div
                            class="clip-hex absolute h-[120px] w-[138px] bg-[#5b3b22]/40"
                            style={`left: ${tile.x - 7}px; top: ${tile.y + 28}px;`}
                        ></div>

                        <div
                            class={[
                                "clip-hex absolute flex h-[108px] w-[124px] items-center justify-center border-[2px] shadow-[0_7px_0_rgba(74,48,31,0.25)]",
                                tile.b === "forest"
                                    ? "border-[#2f6546] bg-[#5b9368]"
                                    : "",
                                tile.b === "mountain"
                                    ? "border-[#656b73] bg-[#a8adb2]"
                                    : "",
                                tile.b === "plain"
                                    ? "border-[#9b7034] bg-[#d9b56a]"
                                    : "",
                                tile.b === "river"
                                    ? "border-[#327b8d] bg-[#6eb8c5]"
                                    : "",
                            ].join(" ")}
                            style={`left: ${tile.x}px; top: ${tile.y + 33}px;`}
                        >
                            <div
                                class="pointer-events-none absolute inset-[5px] clip-hex border border-white/25"
                            ></div>

                            {#if tile.structure}
                                <div
                                    class="relative z-10 grid h-full w-full place-items-center"
                                >
                                    <StructureGlyph
                                        structure={tile.structure}
                                        owner={tile.owner === "blue"
                                            ? "blue"
                                            : "red"}
                                        size="lg"
                                        boxed
                                    />
                                </div>
                            {/if}
                        </div>
                    {/each}
                </div>
            </div>
        </div>

        <div class="mx-auto w-full max-w-xl text-center lg:text-left">
            <div
                class="mb-6 inline-flex items-center gap-3 rounded-2xl bg-[#23444c] px-5 py-3 shadow-md ring-1 ring-[#f8efe0]/10"
            >
                <div
                    class="grid h-12 w-12 place-items-center rounded-xl bg-[#f2c36b] text-2xl font-black text-[#142833] shadow-md"
                >
                    <span class="logo-diamond">◈</span>
                </div>

                <div>
                    <div
                        class="text-2xl font-semibold tracking-tight text-[#fff7e8]"
                    >
                        Frontiers
                    </div>
                    <div
                        class="text-xs font-semibold uppercase tracking-[0.25em] text-[#9fc9c5]"
                    >
                        Online Board Game
                    </div>
                </div>
            </div>

            <h1
                class="text-5xl font-semibold leading-tight tracking-tight text-[#fff7e8] sm:text-6xl"
            >
                Expand influence.
                <br />
                Control the frontier.
            </h1>

            <p
                class="mt-5 max-w-lg text-lg leading-8 text-[#d9e6df] lg:max-w-none"
            >
                A dice-free multiplayer strategy game about drafting, placing
                hexes, building outposts, upgrading cities, crossing rivers, and
                winning territory through influence.
            </p>

            <div class="mt-8 grid gap-4 sm:grid-cols-2">
                <button
                    class="cursor-pointer rounded-2xl bg-[#c96f3d] px-7 py-4 text-xl font-bold text-white shadow-[0_8px_0_rgba(91,48,28,0.55)] transition hover:-translate-y-0.5 hover:bg-[#dc7b45] active:translate-y-1 active:shadow-[0_4px_0_rgba(91,48,28,0.55)] disabled:cursor-not-allowed disabled:opacity-60"
                    type="button"
                    disabled={loading}
                    on:click={onCreateRoom}
                >
                    Create Room
                </button>

                <button
                    class="cursor-pointer rounded-2xl bg-[#f8efe0]/10 px-7 py-4 text-xl font-bold text-[#fff7e8] shadow-[0_8px_0_rgba(0,0,0,0.18)] ring-1 ring-[#f8efe0]/20 transition hover:-translate-y-0.5 hover:bg-[#f8efe0]/16 active:translate-y-1 disabled:cursor-not-allowed disabled:opacity-60"
                    type="button"
                    disabled={loading}
                    on:click={() => (joinOpen = !joinOpen)}
                >
                    Join Room
                </button>
            </div>

            {#if joinOpen}
                <div
                    class="mt-5 rounded-3xl bg-[#23444c] p-4 shadow-md ring-1 ring-[#f8efe0]/10"
                >
                    <label
                        for="room-code"
                        class="mb-2 block text-sm font-bold uppercase tracking-[0.2em] text-[#9fc9c5]"
                    >
                        Enter room code
                    </label>

                    <div class="grid gap-3 sm:grid-cols-[1fr_auto_auto]">
                        <input
                            id="room-code"
                            bind:value={roomCode}
                            class="rounded-2xl border border-[#f8efe0]/20 bg-[#fff7e8] px-4 py-3 font-bold tracking-wider text-[#142833] outline-none placeholder:text-[#8d7e68] focus:ring-4 focus:ring-[#f2c36b]/30"
                            placeholder="abc12345"
                            disabled={loading}
                            on:keydown={(event) => {
                                if (event.key === "Enter") join();
                            }}
                        />

                        <button
                            class="cursor-pointer rounded-2xl bg-[#f2c36b] px-5 py-3 font-bold text-[#142833] transition hover:bg-[#ffd27c] disabled:cursor-not-allowed disabled:opacity-60"
                            type="button"
                            disabled={loading || roomCode.trim().length === 0}
                            on:click={join}
                        >
                            Join
                        </button>

                        <button
                            class="cursor-pointer rounded-2xl bg-[#73c4bd] px-5 py-3 font-bold text-[#102b38] transition hover:bg-[#85d8d1] disabled:cursor-not-allowed disabled:opacity-60"
                            type="button"
                            disabled={loading || roomCode.trim().length === 0}
                            on:click={spectate}
                        >
                            Watch
                        </button>
                    </div>
                </div>
            {/if}

            {#if error}
                <div
                    class="mt-4 rounded-2xl bg-[#b94b3f] px-5 py-3 font-semibold text-white shadow-lg"
                >
                    {error}
                </div>
            {/if}

            <div
                class="mt-8 flex flex-wrap justify-center gap-5 text-sm font-semibold text-[#b9d5d1] lg:justify-start"
            >
                <span>👥 2+ players</span>
                <span>🧭 20–30 min matches</span>
                <span>🏛 Cities, roads, bridges</span>
            </div>
        </div>
    </section>
</AppShell>

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
