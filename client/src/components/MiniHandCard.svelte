<script lang="ts">
    import type { DraftItem } from "../lib/types";
    import { cardClass, cardIcon, cardTitle, cardType } from "../lib/cardInfo";
    import CardTooltip from "./CardTooltip.svelte";

    export let item: DraftItem | null | undefined;
</script>

<CardTooltip {item}>
    <div
        class={[
            "grid h-10 w-10 place-items-center rounded-xl border-2 text-lg font-black shadow-[0_3px_0_rgba(0,0,0,0.18)]",
            cardClass(item),
        ].join(" ")}
        title={`${cardTitle(item)} ${cardType(item)}`}
    >
        {cardIcon(item)}
    </div>
</CardTooltip>
