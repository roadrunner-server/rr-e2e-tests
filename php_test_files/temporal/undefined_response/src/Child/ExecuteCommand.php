<?php

/**
 * This file is part of Temporal package.
 *
 * For the full copyright and license information, please view the LICENSE
 * file that was distributed with this source code.
 */

declare(strict_types=1);

namespace Temporal\Samples\Child;

use Carbon\CarbonInterval;
use Symfony\Component\Console\Input\InputInterface;
use Symfony\Component\Console\Output\OutputInterface;
use Temporal\Client\WorkflowOptions;
use Temporal\Common\RetryOptions;
use Temporal\Samples\SampleUtils\Command;

class ExecuteCommand extends Command
{
    protected const NAME = 'child';
    protected const DESCRIPTION = 'Execute Child\ParentWorkflow';

    public function execute(InputInterface $input, OutputInterface $output)
    {
        $workflow = $this->workflowClient->newWorkflowStub(
            ParentWorkflowInterface::class,
            WorkflowOptions::new()
                ->withWorkflowExecutionTimeout(CarbonInterval::minute(2))
                ->withRetryOptions(
                    RetryOptions::new()
                        ->withMaximumAttempts(3)
                )
        );

        $output->writeln("Starting <comment>ParentWorkflow</comment>... ");

        $run = $this->workflowClient->start($workflow, 'World');

        $output->writeln(
            sprintf(
                'Started: WorkflowID=<fg=magenta>%s</fg=magenta>, RunID=<fg=magenta>%s</fg=magenta>',
                $run->getExecution()->getID(),
                $run->getExecution()->getRunID(),
            )
        );
        $start = microtime(true);

        $output->writeln(sprintf("Result:\n<info>%s</info>", print_r($run->getResult(), true)));

        $output->writeln(sprintf("Elapsed: <info>%s</info>", \round(microtime(true) - $start, 2)));

        return self::SUCCESS;
    }
}