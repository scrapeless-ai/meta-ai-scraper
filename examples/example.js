// Without an argument, submit the sample question. Pass its task ID to read the answer.
const api = "https://api.scrapeless.com/api/v2/scraper";
const payload = {
  actor: "scraper.meta",
  input: { prompt: "Which public museums in Chicago have architecture exhibits?", country: "US" },
};

async function main() {
  const token = process.env.SCRAPELESS_API_TOKEN;
  if (!token) throw new Error("Set SCRAPELESS_API_TOKEN first.");
  if (process.argv.length > 3) throw new Error("Usage: node examples/example.js [task_id]");

  const taskId = process.argv[2];
  const path = taskId === undefined ? "/request" : "/result/" + encodeURIComponent(taskId);
  const response = await fetch(api + path, {
    method: taskId === undefined ? "POST" : "GET",
    headers: { "Content-Type": "application/json", "x-api-token": token },
    body: taskId === undefined ? JSON.stringify(payload) : undefined,
    signal: AbortSignal.timeout(60000),
  });
  if (!response.ok) throw new Error("HTTP " + response.status + ": " + await response.text());
  const data = await response.json();
  if (taskId === undefined) {
    if (!data.task_id) throw new Error("Creation response has no task ID.");
    console.log(data.task_id);
    return;
  }
  if (data.status === "pending" || data.status === "running") {
    console.log(data.status);
    return;
  }
  if (data.status !== "success") throw new Error(data.message || "Task status: " + data.status);

  const result = data.task_result;
  if (typeof result.full_response !== "string") throw new Error("Response has no full_response.");
  console.log(result.full_response);
  for (const source of result.references || []) {
    console.log("- " + source.title + ": " + source.url);
  }
}

main().catch((error) => {
  console.error(error.message);
  process.exitCode = 1;
});
