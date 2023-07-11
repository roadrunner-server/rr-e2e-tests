<?php

/**
 * This file is part of Temporal package.
 *
 * For the full copyright and license information, please view the LICENSE
 * file that was distributed with this source code.
 */

declare(strict_types=1);

namespace Temporal\Tests\Workflow;

use Temporal\Activity\ActivityInterface;
use Temporal\Activity\ActivityMethod;

#[ActivityInterface(prefix: "SignalWorkflow.")]
class SimpleActivity
{
    #[ActivityMethod]
    public function composeStrings(string $greeting, string $name): string
    {
        return $greeting . ' ' . $name;
    }
}