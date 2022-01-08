<?php

declare(strict_types=1);

namespace Arku\Newrelic\Transactions;

final class Segment implements SegmentInterface
{
    private string $name = 'unnamed segment';
    private string $duration = '0';
    private array $meta = [];

    public function getName(): string
    {
        return $this->name;
    }

    public function setName(string $name): SegmentInterface
    {
        $this->name = $name;
        return $this;
    }

    /**
     * @param string|int|float $duration
     * @return SegmentInterface
     */
    public function setDuration($duration): SegmentInterface
    {
        if (!is_numeric($duration)) {
            throw  new \InvalidArgumentException('Time provided for the segment has to be numeric');
        }
        $this->duration = strval($duration * 1000);
        return $this;
    }

    /**
     * @param array $meta
     */
    public function setMeta(array $meta): void
    {
        $this->meta = $meta;
    }

    public function getDuration(): string
    {
        return $this->duration;
    }

    public function getMeta(): array
    {
        return $this->meta;
    }
}