<?php

/**
 * This file is part of Temporal package.
 *
 * For the full copyright and license information, please view the LICENSE
 * file that was distributed with this source code.
 */

declare(strict_types=1);

namespace Temporal\Samples\Child;

use Temporal\Activity;

#[Activity\ActivityInterface(prefix: 'Child.')]
class GreetingActivity
{
    #[Activity\ActivityMethod(name: "ComposeGreeting")]
    public function composeGreeting(string $greeting, string $name): string
    {
        \file_put_contents('runtime/TIMING_ACTIVITY.txt', \date('r') . " before sleep\n", FILE_APPEND);
        \sleep(5);
        \file_put_contents('runtime/TIMING_ACTIVITY.txt', \date('r') . " after sleep\n", FILE_APPEND);

        return $greeting . ' ' . $name;
    }
}