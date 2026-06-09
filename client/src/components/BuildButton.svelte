<script lang="ts">
    import CostBadge from "./CostBadge.svelte";
    import BuildActionTooltip from "./BuildActionTooltip.svelte";
    import type { BuildAction, TargetBuildAction } from "../lib/types";

    type Cost = {
        wood?: number;
        stone?: number;
        grain?: number;
        relic?: number;
    };

    export let action: BuildAction = "outpost";
    export let label = "Build";
    export let icon = "?";
    export let hint = "";
    export let title = "";
    export let cost: Cost = {};
    export let affordable = true;
    export let enabled = true;
    export let selected = false;
    export let variant: "default" | "water" = "default";
    export let tokenText = "";

    export let onSelect: (action: TargetBuildAction) => void = () => {};
    export let onInstant: (action: BuildAction) => void = () => {};

    $: isTargetAction =
        action === "outpost" ||
        action === "settlement" ||
        action === "city" ||
        action === "blockade" ||
        action === "flood";

    $: buttonClass =
        variant === "water"
            ? [
                  "h-full w-full rounded-2xl p-4 text-center font-black shadow-[0_6px_0_rgba(0,0,0,0.18)] transition active:translate-y-1",
                  enabled
                      ? "cursor-pointer bg-[#6eb8c5] text-[#102b38] hover:bg-[#85d8d1]"
                      : "cursor-not-allowed bg-[#f8efe0]/10 text-[#fff7e8] opacity-45 ring-1 ring-[#f8efe0]/20",
              ].join(" ")
            : [
                  "h-full w-full rounded-2xl p-4 text-center font-black shadow-[0_6px_0_rgba(0,0,0,0.18)] transition active:translate-y-1",
                  selected
                      ? "bg-[#f2c36b] text-[#142833]"
                      : "bg-[#f8efe0]/10 text-[#fff7e8] ring-1 ring-[#f8efe0]/20",
                  enabled
                      ? "cursor-pointer hover:bg-[#f8efe0]/16"
                      : "cursor-not-allowed opacity-45",
              ].join(" ");

    function handleClick() {
        if (!enabled) return;

        if (isTargetAction) {
            onSelect(action as TargetBuildAction);
            return;
        }

        onInstant(action);
    }
</script>

<BuildActionTooltip {action} {hint}>
    <button
        class={buttonClass}
        type="button"
        disabled={!enabled}
        {title}
        on:click={handleClick}
    >
        <div class="text-3xl">{icon}</div>

        <div class="mt-1 text-xs uppercase tracking-wider">
            {label}
        </div>

        {#if tokenText}
            <div class="mt-2 text-xs font-black">
                {tokenText}
            </div>
        {:else}
            <div class="mt-2">
                <CostBadge
                    wood={cost.wood ?? 0}
                    stone={cost.stone ?? 0}
                    grain={cost.grain ?? 0}
                    relic={cost.relic ?? 0}
                    {affordable}
                />
            </div>
        {/if}
    </button>
</BuildActionTooltip>
