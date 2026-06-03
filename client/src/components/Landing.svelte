<script lang="ts">
    export type LandingUser = {
        authenticated: boolean;
        isGuest: boolean;
        displayName: string;
        avatarUrl?: string;
        karma: number;
    } | null;

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

    const previewTiles = [
        { b: "mountain", icon: "⛰️", x: 82, y: 0 },
        { b: "plain", icon: "🌾", x: 200, y: 0 },
        { b: "forest", icon: "🌲", x: 318, y: 0 },

        { b: "plain", icon: "🌾", x: 23, y: 88 },
        { b: "forest", icon: "🌲", x: 141, y: 88 },
        { b: "river", icon: "💧", x: 259, y: 88 },
        { b: "mountain", icon: "⛰️", x: 377, y: 88 },

        { b: "forest", icon: "🌲", x: 82, y: 176 },
        { b: "plain", icon: "🌾", x: 200, y: 176 },
        { b: "mountain", icon: "⛰️", x: 318, y: 176 },

        { b: "river", icon: "💧", x: 23, y: 264 },
        { b: "plain", icon: "🌾", x: 141, y: 264 },
        { b: "forest", icon: "🌲", x: 259, y: 264 },
        { b: "plain", icon: "🌾", x: 377, y: 264 },
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

<main
    class="relative min-h-screen overflow-hidden bg-[#0e2430] font-sans text-[#f8efe0]"
>
    <!-- Background -->
    <div class="pointer-events-none absolute inset-0">
        <div
            class="absolute inset-0 bg-[radial-gradient(circle_at_20%_20%,rgba(217,160,91,0.18),transparent_32%),radial-gradient(circle_at_80%_10%,rgba(99,179,177,0.16),transparent_30%),linear-gradient(135deg,#102b38_0%,#0b1d27_52%,#132f32_100%)]"
        ></div>

        <div class="absolute inset-0 opacity-[0.055]">
            <div class="map-grid h-full w-full"></div>
        </div>

        <div
            class="absolute -left-24 top-20 h-72 w-72 rounded-full bg-[#d9a05b]/20 blur-3xl"
        ></div>
        <div
            class="absolute bottom-10 right-10 h-80 w-80 rounded-full bg-[#4fb4ae]/15 blur-3xl"
        ></div>
    </div>

    <!-- Header -->
    <header
        class="relative z-10 mx-auto flex max-w-7xl items-center justify-between px-6 py-5 lg:px-12"
    >
        <div class="flex items-center gap-3">
            <div
                class="grid h-11 w-11 place-items-center rounded-2xl bg-[#f2c36b] text-xl font-black text-[#142833] shadow-[0_8px_0_rgba(0,0,0,0.16)] ring-1 ring-white/20"
            >
                ◈
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
                    Tactical Map Game
                </div>
            </div>
        </div>

        <div class="flex items-center gap-3">
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
                    class="cursor-pointer rounded-xl px-4 py-2 text-sm font-semibold text-[#f8efe0] transition hover:bg-white/10"
                    type="button"
                    on:click={onLogin}
                >
                    Log in
                </button>

                <button
                    class="cursor-pointer rounded-xl bg-[#f8efe0] px-4 py-2 text-sm font-bold text-[#142833] shadow-sm transition hover:bg-white"
                    type="button"
                    on:click={onSignUp}
                >
                    Sign up
                </button>
            {/if}
        </div>
    </header>

    <!-- Hero -->
    <section
        class="relative z-10 mx-auto grid min-h-[calc(100vh-84px)] max-w-7xl grid-cols-1 items-center gap-10 px-6 pb-12 lg:grid-cols-[1.05fr_0.95fr] lg:px-12"
    >
        <!-- Board preview -->
        <div class="flex justify-center lg:justify-start">
            <div class="relative h-[455px] w-[610px] max-w-full">
                <div
                    class="absolute inset-0 rounded-[48px] bg-[#f2c36b]/15 blur-3xl"
                ></div>

                <div
                    class="absolute left-1/2 top-1/2 h-[390px] w-[555px] -translate-x-1/2 -translate-y-1/2 rotate-[-2deg] rounded-[34px] bg-[#d8b985] shadow-2xl ring-1 ring-black/20"
                ></div>

                <div
                    class="absolute left-1/2 top-1/2 h-[360px] w-[525px] -translate-x-1/2 -translate-y-1/2 rotate-[-2deg] rounded-[28px] border border-[#5d4128]/25 bg-[#efe0bd] shadow-inner"
                >
                    <div
                        class="absolute inset-0 rounded-[28px] opacity-25 map-paper"
                    ></div>
                </div>

                <div class="relative mx-auto h-[420px] w-[560px] translate-y-8">
                    {#each previewTiles as tile}
                        <!-- small underlay removes visual gaps and gives tile depth -->
                        <div
                            class="clip-hex absolute h-[120px] w-[138px] bg-[#5b3b22]/45 shadow-lg"
                            style={`left: ${tile.x - 7}px; top: ${tile.y - 5}px;`}
                        ></div>

                        <div
                            class={[
                                "clip-hex absolute flex h-[108px] w-[124px] items-center justify-center border-[2px] shadow-[0_10px_0_rgba(74,48,31,0.22),0_18px_28px_rgba(30,23,15,0.20)]",
                                tile.b === "forest"
                                    ? "border-[#28583f] bg-[#4f8f5f]"
                                    : "",
                                tile.b === "mountain"
                                    ? "border-[#5a6170] bg-[#a8afb8]"
                                    : "",
                                tile.b === "plain"
                                    ? "border-[#9b7034] bg-[#ddb769]"
                                    : "",
                                tile.b === "river"
                                    ? "border-[#2d7286] bg-[#65b7c7]"
                                    : "",
                            ].join(" ")}
                            style={`left: ${tile.x}px; top: ${tile.y}px;`}
                        >
                            <div
                                class="pointer-events-none absolute inset-[5px] clip-hex border border-white/30"
                            ></div>
                            <div
                                class="pointer-events-none absolute inset-0 bg-[radial-gradient(circle_at_30%_25%,rgba(255,255,255,0.26),transparent_34%)]"
                            ></div>

                            <div class="grid h-full w-full place-items-center">
                                <div class="text-4xl drop-shadow-sm">
                                    {tile.icon}
                                </div>
                            </div>

                            <!-- Fixed coordinates -->
                            {#if tile.x === 200 && tile.y === 176}
                                <div
                                    class="absolute -bottom-2 left-1/2 h-11 w-9 -translate-x-1/2 rounded-t-full bg-[#1d4e89] shadow-lg ring-2 ring-[#fff7e8]"
                                ></div>
                            {/if}

                            {#if tile.x === 259 && tile.y === 264}
                                <div
                                    class="absolute -bottom-2 left-1/2 h-11 w-9 -translate-x-1/2 rounded-t-full bg-[#b94b3f] shadow-lg ring-2 ring-[#fff7e8]"
                                ></div>
                            {/if}
                        </div>
                    {/each}
                </div>

                <div
                    class="absolute bottom-3 left-10 rounded-full bg-[#0e2430]/80 px-4 py-2 text-xs font-bold uppercase tracking-[0.2em] text-[#f2c36b] ring-1 ring-[#f2c36b]/30 backdrop-blur"
                >
                    Live map preview
                </div>
            </div>
        </div>

        <!-- Text / controls -->
        <div class="mx-auto w-full max-w-xl text-center lg:text-left">
            <div
                class="mb-6 inline-flex items-center gap-3 rounded-2xl bg-[#f8efe0]/10 px-5 py-3 shadow-lg ring-1 ring-[#f8efe0]/15 backdrop-blur"
            >
                <div
                    class="grid h-12 w-12 place-items-center rounded-xl bg-[#f2c36b] text-2xl font-black text-[#142833] shadow-md"
                >
                    ◈
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
                Command the map.
                <br />
                Claim the edge.
            </h1>

            <p
                class="mt-5 max-w-lg text-lg leading-8 text-[#d9e6df] lg:max-w-none"
            >
                A fast browser strategy game about routes, settlements, rivers,
                resources, and pressure across a changing frontier.
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
                    class="mt-5 rounded-3xl bg-[#f8efe0]/10 p-4 shadow-xl ring-1 ring-[#f8efe0]/15 backdrop-blur"
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
                <span>👥 2 players + spectators</span>
                <span>🧭 20–30 min matches</span>
                <span>🗺️ Fresh tactical maps</span>
            </div>
        </div>
    </section>
</main>

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

    .map-grid {
        background-image:
            linear-gradient(rgba(255, 255, 255, 0.18) 1px, transparent 1px),
            linear-gradient(
                90deg,
                rgba(255, 255, 255, 0.18) 1px,
                transparent 1px
            );
        background-size: 72px 72px;
    }

    .map-paper {
        background-image:
            radial-gradient(
                circle at 20% 30%,
                rgba(91, 65, 40, 0.22) 0 1px,
                transparent 2px
            ),
            radial-gradient(
                circle at 80% 40%,
                rgba(91, 65, 40, 0.18) 0 1px,
                transparent 2px
            ),
            radial-gradient(
                circle at 45% 70%,
                rgba(91, 65, 40, 0.16) 0 1px,
                transparent 2px
            );
        background-size:
            42px 42px,
            58px 58px,
            70px 70px;
    }
</style>
