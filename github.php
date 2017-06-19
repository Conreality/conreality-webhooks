<?php
/**
 * Post-push webhook for project GitHub repositories.
 *
 * See: https://github.com/conreality/prov.conreality.org/settings/hooks
 * See: https://twitter.com/ConrealityProv
 * See: https://developer.github.com/webhooks/
 * See: https://support.twitter.com/articles/78124
 */

require __DIR__ . '/vendor/autoload.php';

function fail_and_die($reason) {
  http_response_code(500);
  header('Content-Type: text/plain; charset=UTF-8');
  die("500 Internal Server Error ($reason)");
}

$secret_file = __DIR__ . '/../../.secret/twitter.conrealityprov.json';
if (!file_exists($secret_file)) {
  fail_and_die("Missing credentials file");
}

$secret = json_decode(file_get_contents($secret_file));
if (empty($secret->consumer_key) || empty($secret->consumer_secret) ||
    empty($secret->access_token) || empty($secret->access_token_tecret)) {
  fail_and_die("Missing credentials data");
}

$push = file_get_contents('php://input');
//file_put_contents('.push.json', $push);  // DEBUG
//$push = file_get_contents('.push.json'); // DEBUG
$push = json_decode($push);

$repository_name = $push->repository->name;
$pusher_name     = $push->pusher->name; // good enough for now
$commit          = $push->head_commit;
$commit_sha1     = $commit->id;
$commit_url      = $commit->url;
$commit_text     = $commit->message;

if (strpos($commit_text, "\n") !== false) {
  $commit_text = substr($commit_text, 0, strpos($commit_text, "\n"));
}

if (strlen($commit_text) > 58) { // 40+1+2+1+12+2+58+1+23 = 140
  $commit_text = substr($commit_text, 0, 58 - 1) . "\u{2026}";
}

$log_message =  "$commit_sha1 by $pusher_name: $commit_text\n$commit_url";
//var_dump($log_message); die(); // DEBUG

$twitter = new Twitter($secret->consumer_key, $secret->consumer_secret,
  $secret->access_token, $secret->access_token_tecret);

try {
  $twitter->send($log_message);
}
catch (TwitterException $error) {
  fail_and_die('Twitter API: ' . $error->getMessage());
}

http_response_code(202);
header('Content-Type: text/plain; charset=UTF-8');
print($log_message);
