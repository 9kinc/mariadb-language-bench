<?php
declare(strict_types=1);
if ($_SERVER['REQUEST_METHOD'] !== 'GET' || parse_url($_SERVER['REQUEST_URI'], PHP_URL_PATH) !== '/lookup') {
    http_response_code(404); exit;
}
$id = filter_input(INPUT_GET, 'id', FILTER_VALIDATE_INT, ['options' => ['min_range' => 1, 'max_range' => 1000000]]);
if ($id === false || $id === null) {http_response_code(400); echo 'invalid id'; exit;}
const SQL = 'SELECT u.id,u.username,u.balance_cents,p.bio,o.amount_cents FROM users AS u INNER JOIN profiles AS p ON p.user_id=u.id INNER JOIN orders AS o ON o.user_id=u.id WHERE u.id = ? LIMIT 1';
try {
    $pdo = new PDO(
      'mysql:host='.(getenv('DB_HOST') ?: '127.0.0.1').';port=3306;dbname=bench;charset=utf8mb4',
      getenv('DB_USER') ?: 'bench',
      getenv('DB_PASSWORD') ?: '',
      [
        PDO::ATTR_ERRMODE => PDO::ERRMODE_EXCEPTION,
        PDO::ATTR_EMULATE_PREPARES => false,
        PDO::ATTR_PERSISTENT => true,
        PDO::ATTR_TIMEOUT => 5,
        PDO::ATTR_STRINGIFY_FETCHES => false
      ]
    );
    $stmt = $pdo->prepare(SQL);
    $stmt->execute([$id]);
    $record = $stmt->fetch(PDO::FETCH_ASSOC);
    if ($record === false) {http_response_code(404);exit;}
    header('Content-Type: application/json');
    echo json_encode($record, JSON_THROW_ON_ERROR);
} catch (Throwable $e) {
    error_log('database read failed: '.$e->getMessage());
    http_response_code(503); echo 'database unavailable';
}
