<?php

namespace Arku\Newrelic\Transactions;

interface SegmentInterface
{
    public function getName(): string;

    public function setName(string $name): SegmentInterface;

    /**
     * @param string|int|float $duration
     * @return SegmentInterface
     */
    public function setDuration($duration): SegmentInterface;

    /**
     * @param array $meta
     */
    public function setMeta(array $meta): void;

    public function getDuration(): string;

    public function getMeta(): array;
}