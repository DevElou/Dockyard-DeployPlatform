const cards = [
  {
    title: "Control Plane",
    text: "API Go responsable de l'etat canonique, des commandes metier et de l'exposition HTTP.",
  },
  {
    title: "Orchestrator Worker",
    text: "Worker Go charge des builds, releases, deployments et rollbacks asynchrones.",
  },
  {
    title: "Deploy Agent",
    text: "Agent Go leger installe sur chaque host Docker pour appliquer un DeploymentSpec.",
  },
];

export default function HomePage() {
  return (
    <main className="page-shell">
      <section className="hero">
        <p className="eyebrow">Dockyard</p>
        <h1>Private deployment control plane for Docker-first infrastructure.</h1>
        <p className="lede">
          V1 cible: Next.js pour l&apos;interface, Go pour l&apos;API, le worker et l&apos;agent,
          avec une separation nette entre metier, orchestration et execution runtime.
        </p>
      </section>

      <section className="grid">
        {cards.map((card) => (
          <article className="card" key={card.title}>
            <h2>{card.title}</h2>
            <p>{card.text}</p>
          </article>
        ))}
      </section>
    </main>
  );
}
