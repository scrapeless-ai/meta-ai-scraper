// Java 17+ and jq. Run without arguments to submit, or pass an existing task ID.
import java.net.URI;
import java.net.URLEncoder;
import java.net.http.HttpClient;
import java.net.http.HttpRequest;
import java.net.http.HttpResponse;
import java.nio.charset.StandardCharsets;
import java.time.Duration;

public class Example {
    private static final String API = "https://api.scrapeless.com/api/v2/scraper";
    private static final String PAYLOAD = """
        {"actor":"scraper.meta","input":{"prompt":"Which public museums in Chicago have architecture exhibits?","country":"US"}}
        """;
    private static final String RESULT_FILTER = """
        if .status == "success" then
          .task_result | (.full_response // error("Response has no full_response.")),
          ((.references // [])[] | "- " + .title + ": " + .url)
        elif .status == "pending" or .status == "running" then .status
        else error(.message // "Task failed.") end
        """;

    public static void main(String[] args) throws Exception {
        String token = System.getenv("SCRAPELESS_API_TOKEN");
        if (token == null || token.isBlank() || args.length > 1) {
            throw new IllegalArgumentException("Set SCRAPELESS_API_TOKEN and run java examples/Example.java [task_id].");
        }
        boolean create = args.length == 0;
        String path = create ? "/request" : "/result/" + URLEncoder.encode(args[0], StandardCharsets.UTF_8);
        HttpRequest.Builder request = HttpRequest.newBuilder(URI.create(API + path))
                .timeout(Duration.ofSeconds(60))
                .header("Content-Type", "application/json")
                .header("x-api-token", token);
        if (create) {
            request.POST(HttpRequest.BodyPublishers.ofString(PAYLOAD));
        }
        HttpResponse<String> response = HttpClient.newHttpClient().send(
                request.build(), HttpResponse.BodyHandlers.ofString(StandardCharsets.UTF_8));
        if (response.statusCode() < 200 || response.statusCode() >= 300) {
            throw new IllegalStateException("HTTP " + response.statusCode() + ": " + response.body());
        }
        String filter = create ? ".task_id // error(\"Creation response has no task ID.\")" : RESULT_FILTER;
        Process jq = new ProcessBuilder("jq", "-er", filter)
                .redirectOutput(ProcessBuilder.Redirect.INHERIT)
                .redirectError(ProcessBuilder.Redirect.INHERIT)
                .start();
        try (var input = jq.getOutputStream()) {
            input.write(response.body().getBytes(StandardCharsets.UTF_8));
        }
        System.exit(jq.waitFor());
    }
}
