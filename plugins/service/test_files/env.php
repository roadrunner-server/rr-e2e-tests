<?php
// foo -> bar
// bar -> foo
//var_dump($_SERVER);
if ($_SERVER["foo"] !== "BAR" ) {
	die("faillll");
}

if ($_SERVER["bar"] !== "FOO" ) {
	die("faillll");
}

exit();