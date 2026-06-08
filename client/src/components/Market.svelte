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

    type VisualSlot = {
        item: DraftItem;
        marketIndex: number;
    } | null;

    let visualSlots: VisualSlot[] = Array(MARKET_SIZE).fill(null);
    let lastMarketSignature = "";

    $: me = game.Players.find((p) => p.Id === playerId);
    $: hand = me?.Hand ?? [];
    $: handCount = hand.length;
    $: remainingPicks = Math.max(0, MAX_HAND - handCount);

    $: canPick =
        role === "player" &&
        game.CurrentPhase === GamePhase.Pick &&
        game.CurrentPlayer === playerId &&
        handCount < MAX_HAND;

    $: {
        const signature = `${game.Round}|${game.CurrentPhase}|${game.CurrentPlayer}|${marketSignature(game.Market)}`;

        if (signature !== lastMarketSignature) {
            visualSlots = reconcileSlots(visualSlots, game.Market);
            lastMarketSignature = signature;
        }
    }

    function marketSignature(market: DraftItem[]) {
        return market.map(itemSignature).join("|");
    }

    function itemSignature(item: DraftItem) {
        return `${item.Kind}:${item.Biome}:${item.Structure}:${item.Action}`;
    }

    function sameItem(a: DraftItem, b: DraftItem) {
        return itemSignature(a) === itemSignature(b);
    }

    function isEmptyVisualSlots(slots: VisualSlot[]) {
        return slots.every((slot) => slot === null);
    }

    function resetSlots(market: DraftItem[]): VisualSlot[] {
        return Array.from({ length: MARKET_SIZE }, (_, index) => {
            const item = market[index];
            if (!item) return null;

            return {
                item,
                marketIndex: index,
            };
        });
    }

    function reconcileSlots(
        previousSlots: VisualSlot[],
        nextMarket: DraftItem[],
    ): VisualSlot[] {
        if (nextMarket.length === 0) {
            return Array(MARKET_SIZE).fill(null);
        }

        if (isEmptyVisualSlots(previousSlots)) {
            return resetSlots(nextMarket);
        }

        const nextSlots: VisualSlot[] = Array(MARKET_SIZE).fill(null);

        const remaining = nextMarket.map((item, marketIndex) => ({
            item,
            marketIndex,
        }));

        for (let slotIndex = 0; slotIndex < MARKET_SIZE; slotIndex++) {
            const previous = previousSlots[slotIndex];
            if (!previous) continue;

            const foundIndex = remaining.findIndex((entry) =>
                sameItem(entry.item, previous.item),
            );

            if (foundIndex >= 0) {
                const [entry] = remaining.splice(foundIndex, 1);
                nextSlots[slotIndex] = entry;
            }
        }

        for (let slotIndex = 0; slotIndex < MARKET_SIZE; slotIndex++) {
            if (nextSlots[slotIndex]) continue;
            if (remaining.length === 0) break;

            const entry = remaining.shift();
            nextSlots[slotIndex] = entry ?? null;
        }

        return nextSlots;
    }

    function pick(slotIndex: number, slot: Exclude<VisualSlot, null>) {
        debugLog("market.pick.click", {
            slotIndex,
            marketIndex: slot.marketIndex,
            canPick,
            role,
            playerId,
            currentPlayer: game.CurrentPlayer,
            currentPhase: game.CurrentPhase,
            handCount,
            marketSize: game.Market.length,
            item: slot.item,
        });

        if (!canPick) return;

        onPick(slot.marketIndex);
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
                {game.Market.length} left
            {:else}
                {game.Market.length} cards
            {/if}
        </div>
    </div>

    <div
        class="mt-4 grid grid-cols-2 gap-3 rounded-3xl bg-[#142833]/35 p-3 ring-1 ring-[#f8efe0]/10"
    >
        {#each visualSlots as slot, slotIndex (slotIndex)}
            <div
                class="relative h-32 rounded-2xl bg-[#102832] p-1 shadow-[inset_0_3px_8px_rgba(0,0,0,0.35),0_1px_0_rgba(255,255,255,0.05)] ring-1 ring-[#f8efe0]/10"
            >
                {#if slot}
                    <CardTooltip
                        item={slot.item}
                        hint={canPick ? "Click to draft" : ""}
                    >
                        <button
                            class={[
                                "block h-full w-full rounded-2xl text-left transition",
                                canPick
                                    ? "cursor-pointer hover:-translate-y-1 hover:brightness-105 active:translate-y-1"
                                    : "cursor-not-allowed opacity-60",
                            ].join(" ")}
                            type="button"
                            disabled={!canPick}
                            on:click={() => pick(slotIndex, slot)}
                        >
                            <DraftCard item={slot.item} disabled={!canPick} />
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
