<script lang="ts">
    import type { DraftItem } from "../lib/types";
    import { cardClass, cardIcon, cardTitle, cardType } from "../lib/cardInfo";
    import CardTooltip from "./CardTooltip.svelte";

    export let item: DraftItem | null | undefined;
    export let size: "sm" | "md" | "lg" = "sm";

    $: sizeClass =
        size === "lg" ? "h-32 p-4" : size === "md" ? "h-24 p-3" : "h-16 p-2";

    $: iconClass =
        size === "lg"
            ? "h-14 w-14 text-3xl rounded-2xl"
            : size === "md"
              ? "h-11 w-11 text-2xl rounded-xl"
              : "h-9 w-9 text-xl rounded-xl";

    $: titleClass =
        size === "lg" ? "text-xl" : size === "md" ? "text-base" : "text-sm";
</script>

<CardTooltip {item}>
    <div
        class={[
            "relative overflow-hidden rounded-2xl border-2 shadow-[0_5px_0_rgba(0,0,0,0.18)]",
            sizeClass,
            cardClass(item),
        ].join(" ")}
        title={`${cardTitle(item)} ${cardType(item)}`}
    >
        <div class="flex h-full items-center gap-3">
            <div
                class={[
                    "grid shrink-0 place-items-center bg-white/35 font-black",
                    iconClass,
                ].join(" ")}
            >
                {cardIcon(item)}
            </div>

            <div class="min-w-0">
                <div
                    class="text-[9px] font-black uppercase tracking-wider opacity-70"
                >
                    {cardType(item)}
                </div>

                <div
                    class={[
                        "truncate font-black leading-tight",
                        titleClass,
                    ].join(" ")}
                >
                    {cardTitle(item)}
                </div>
            </div>
        </div>
    </div>
</CardTooltip>
