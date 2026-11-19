<?php
// Requires PHP 8+ with the OpenSSL extension and HTTPS streams enabled.
declare(strict_types=1);

$token = getenv('SCRAPELESS_API_TOKEN');
if (!$token || $argc > 2) {
    fwrite(STDERR, "Set SCRAPELESS_API_TOKEN and run php examples/example.php [task_id].\n");
    exit(1);
}

$create = $argc === 1;
$path = $create ? '/request' : '/result/' . rawurlencode($argv[1]);
$options = [
    'method' => $create ? 'POST' : 'GET',
    'header' => ["Content-Type: application/json", "x-api-token: " . $token],
    'timeout' => 60,
    'follow_location' => 0,
];

if ($create) {
    $options['content'] = json_encode([
        'actor' => 'scraper.meta',
        'input' => ['prompt' => 'Which public museums in Chicago have architecture exhibits?', 'country' => 'US'],
    ], JSON_THROW_ON_ERROR);
}

$url = 'https://api.scrapeless.com/api/v2/scraper' . $path;
$body = file_get_contents($url, false, stream_context_create(['http' => $options]));
if ($body === false) {
    exit(1);
}
$data = json_decode($body, true, 512, JSON_THROW_ON_ERROR);
if ($create) {
    if (empty($data['task_id'])) {
        fwrite(STDERR, "Creation response has no task ID.\n");
        exit(1);
    }
    echo $data['task_id'], PHP_EOL;
    exit(0);
}

$status = $data['status'] ?? 'unknown';
if (in_array($status, ['pending', 'running'], true)) {
    echo $status, PHP_EOL;
    exit(0);
}
if ($status !== 'success' || !isset($data['task_result']['full_response'])) {
    fwrite(STDERR, ($data['message'] ?? "No answer returned (status: $status).") . PHP_EOL);
    exit(1);
}
$result = $data['task_result'];
echo $result['full_response'], PHP_EOL;
foreach ($result['references'] ?? [] as $source) {
    echo '- ', $source['title'], ': ', $source['url'], PHP_EOL;
}
