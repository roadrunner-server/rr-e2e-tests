<?php

namespace Arku\Newrelic\Transformers;

use Arku\Newrelic\Transactions\SegmentInterface;
use Arku\Newrelic\Transactions\TransactionDetailInterface;

interface TransactionDetailTransformerInterface
{
    public const THROWABLE_MESSAGE = 'message';
    public const THROWABLE_CLASS = 'class';
    public const THROWABLE_STACKTRACE = 'attributes';

    public function transform(TransactionDetailInterface $transactionDetail): array;

    public function transformSegment(SegmentInterface $segment): array;

    public function transformThrowable(\Throwable $throwable): array;
}