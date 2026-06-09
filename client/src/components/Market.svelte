<script lang="ts">
    import DraftCard from "./DraftCard.svelte";
    import CardTooltip from "./CardTooltip.svelte";
    import { GamePhase, type DraftItem, type GameState } from "../lib/types";
    import { debugLog } from "../lib/debug";

    export let game: GameState;
    export let playerId = 0;
    export let role: "player" | "spectator" | "" = "";

    export let onPick: (marketIndex: number) => void;

    const MARKET_SIZE = 6;
    const MAX_HAND = 3;

    $: me = game.Players.find((p) => p.Id === playerId);
    $: hand = me?.Hand ?? [];
    $: handCount = hand.length;
    $: remainingPicks = Math.max(0, MAX_HAND - handCount);

    $: marketSlots = Array.from(
        { length: MARKET_SIZE },
        (_, index) => game.Market[index] ?? null,
    );

    $: filledCount = marketSlots.filter((slot) => slot !== null).length;

    $: canPick =
        role === "player" &&
        game.CurrentPhase === GamePhase.Pick &&
        game.CurrentPlayer === playerId &&
        handCount < MAX_HAND;

    function pick(slotIndex: number, item: DraftItem) {
        debugLog("market.pick.click", {
            slotIndex,
            sentMarketIndex: slotIndex,
            canPick,
            role,
            playerId,
            currentPlayer: game.CurrentPlayer,
            currentPhase: game.CurrentPhase,
            handCount,
            marketRaw: game.Market,
            marketSlots,
            item,
        });

        if (!canPick) return;

        // Send the exact backend slot index.
        onPick(slotIndex);
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
            {#if canPick}
                Pick {remainingPicks}
            {:else if game.CurrentPhase === GamePhase.Pick}
                {filledCount}/6 left
            {:else}
                {filledCount}/6 cards
            {/if}
        </div>
    </div>

    <div
        class="mt-4 grid grid-cols-2 gap-3 rounded-3xl bg-[#142833]/35 p-3 ring-1 ring-[#f8efe0]/10"
    >
        {#each marketSlots as item, slotIndex (slotIndex)}
            <div
                class="h-32 rounded-2xl bg-[#102832] p-1 shadow-[inset_0_3px_8px_rgba(0,0,0,0.35),0_1px_0_rgba(255,255,255,0.05)] ring-1 ring-[#f8efe0]/10"
            >
                {#if item}
                    <CardTooltip {item} hint={canPick ? "Click to draft" : ""}>
                        <button
                            class={[
                                "block h-full w-full rounded-2xl text-left transition",
                                canPick
                                    ? "cursor-pointer hover:-translate-y-1 hover:brightness-105 active:translate-y-1"
                                    : "cursor-not-allowed opacity-60",
                            ].join(" ")}
                            type="button"
                            disabled={!canPick}
                            on:click={() => pick(slotIndex, item)}
                        >
                            {#key `${slotIndex}-${item.Kind}-${item.Biome}-${item.Structure}-${item.Action}`}
                                <DraftCard {item} disabled={!canPick} />
                            {/key}
                        </button>
                    </CardTooltip>
                {:else}
                    <div
                        class="market-empty-slot grid h-full place-items-center rounded-2xl border-2 border-dashed border-[#f8efe0]/16 bg-[#f8efe0]/5 text-center"
                    >
                        <div>
                            <div
                                class="mx-auto grid h-10 w-10 place-items-center rounded-xl bg-[#f8efe0]/8 text-2xl font-black text-[#9fc9c5]/70"
                            >
                                —
                            </div>

                            <div
                                class="mt-3 text-[10px] font-black uppercase tracking-[0.2em] text-[#9fc9c5]/70"
                            >
                                Empty
                            </div>
                        </div>
                    </div>
                {/if}
            </div>
        {/each}
    </div>
</section>

<style>
    .market-empty-slot {
        animation: empty-slot-in 180ms ease-out;
    }

    @keyframes empty-slot-in {
        0% {
            transform: scale(0.96);
            opacity: 0;
        }

        100% {
            transform: scale(1);
            opacity: 1;
        }
    }
</style>
