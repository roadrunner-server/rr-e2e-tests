<?php

namespace Arku\Newrelic\Transactions;

use Throwable;

interface TransactionDetailInterface
{
    public function setName(string $name): TransactionDetailInterface;

    public function setCustomData(string $key, string $value): TransactionDetailInterface;

    /**
     * @return string[]
     */
    public function getDetailsAsArray(): array;

    public function addSegment(SegmentInterface $segment): self;

    public function getSegments(): array;

    public function setThrowable(Throwable $throwable): self;
    public function getThrowable(): ?Throwable;

}