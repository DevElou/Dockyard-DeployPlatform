// Maps backend phase keys to short, human-readable labels for the UI timeline.
// Unknown phases fall back to a humanized form (snake_case → Title Case).
const KNOWN_PHASES: Record<string, string> = {
  // Release / build
  queued: "Queued",
  resolving_source: "Resolving source",
  downloading_archive: "Downloading source",
  building_image: "Building image",
  pushing_image: "Pushing image",
  inspecting_digest: "Reading image digest",
  succeeded: "Succeeded",

  // Deployment
  building_spec: "Building deployment spec",
  contacting_agent: "Contacting deploy agent",
  pulling_image: "Pulling image",
  starting_container: "Starting container",
  health_check: "Waiting for healthy",
  routing: "Configuring routing",
  healthy: "Healthy",

  // Shared
  failed: "Failed",
};

export function humanizePhase(phase: string): string {
  if (KNOWN_PHASES[phase]) return KNOWN_PHASES[phase];
  return phase
    .split("_")
    .filter(Boolean)
    .map((w) => w.charAt(0).toUpperCase() + w.slice(1))
    .join(" ");
}
