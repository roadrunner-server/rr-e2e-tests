<?php

/**
 * This file is part of Temporal package.
 *
 * For the full copyright and license information, please view the LICENSE
 * file that was distributed with this source code.
 */

declare(strict_types=1);

namespace Temporal\Samples\Child;

use Temporal\Common\RetryOptions;
use Temporal\Workflow;

/**
 * Demonstrates a child workflow. Requires a local instance of the Temporal server to be running.
 */
class ParentWorkflow implements ParentWorkflowInterface
{
    public function greet(string $name)
    {
        \file_put_contents('runtime/TIMING_PARENT.txt', \date('r') . " Started\n", FILE_APPEND);
        $child = Workflow::newChildWorkflowStub(
            ChildWorkflow::class,
            Workflow\ChildWorkflowOptions::new()
                ->withParentClosePolicy(Workflow\ParentClosePolicy::POLICY_ABANDON)
                ->withWorkflowExecutionTimeout(20)
                ->withWorkflowTaskTimeout(20)
                ->withRetryOptions(
                    RetryOptions::new()
                        ->withMaximumAttempts(3)
                )
        );

        // This is a non blocking call that returns immediately.
        // Use yield $child->greet(name) to call synchronously.
        $childGreet = yield $child->greet($name);
        \file_put_contents('runtime/TIMING_PARENT.txt', \date('r') . " Greeted\n", FILE_APPEND);

        // Do something else here.

        return 'Hello ' . $name . ' from parent; ' . $childGreet;
    }
}