<script lang="ts" context="module">
    import type { Resource as ResourceType } from "../lib/types";

    export type ResourceFlight = {
        id: number;
        resource: ResourceType;
        amount: number;
        fromX: number;
        fromY: number;
        toX: number;
        toY: number;
        label?: string;
    };
</script>

<script lang="ts">
    import ResourceIcon from "./ResourceIcon.svelte";
    import { Resource } from "../lib/types";

    export let flights: ResourceFlight[] = [];

    function resourceName(
        resource: Resource,
    ): "wood" | "stone" | "grain" | "relic" {
        switch (resource) {
            case Resource.Wood:
                return "wood";
            case Resource.Stone:
                return "stone";
            case Resource.Grain:
                return "grain";
            case Resource.Relic:
                return "relic";
            default:
                return "wood";
        }
    }
</script>

<div class="pointer-events-none fixed inset-0 z-[2147483000] overflow-hidden">
    {#each flights as flight (flight.id)}
        <div
            class="resource-flight"
            style={`--from-x: ${flight.fromX}px; --from-y: ${flight.fromY}px; --to-x: ${flight.toX}px; --to-y: ${flight.toY}px;`}
        >
            <ResourceIcon
                resource={resourceName(flight.resource)}
                amount={flight.label ?? `+${flight.amount}`}
                size="md"
                pulse
            />
        </div>
    {/each}
</div>

<style>
    .resource-flight {
        position: fixed;
        left: 0;
        top: 0;
        animation: resource-flight 1550ms cubic-bezier(0.16, 1, 0.3, 1) forwards;
        filter: drop-shadow(0 10px 12px rgba(0, 0, 0, 0.28));
        will-change: transform, opacity;
    }

    @keyframes resource-flight {
        0% {
            transform: translate(var(--from-x), var(--from-y))
                translate(-50%, -50%) scale(0.75);
            opacity: 0;
        }

        12% {
            opacity: 1;
        }

        55% {
            transform: translate(
                    calc((var(--from-x) + var(--to-x)) / 2),
                    calc((var(--from-y) + var(--to-y)) / 2 - 48px)
                )
                translate(-50%, -50%) scale(1.08);
            opacity: 1;
        }

        100% {
            transform: translate(var(--to-x), var(--to-y)) translate(-50%, -50%)
                scale(0.58);
            opacity: 0;
        }
    }
</style>
