<?php
// foo -> bar
// bar -> foo
//var_dump($_SERVER);
if ($_SERVER["FOO"] !== "BAR" ) {
	die("faillll");
}

if ($_SERVER["BAR"] !== "FOO" ) {
	die("faillll");
}

exit();