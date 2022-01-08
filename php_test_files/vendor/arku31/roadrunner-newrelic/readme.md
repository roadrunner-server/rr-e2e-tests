#Response enricher for the newrelic implementation in the roadrunner


Usage


    $transactionDetail = new TransactionDetail();
    $transactionDetail->setName('test');
    $transactionDetail->setCustomData('key', 'value');

    $segment = new Segment();
    $segment->setName('testSegment');
    $segment->setDuration('1');
    $segment->setMeta(['testmetakey' => 'testmetavalue']);

    $transactionDetail->addSegment($segment);

    $enricher = new EnrichResponse(new TransactionDetailTransformer());
    $response = $enricher->enrich($response, $transactionDetail);


## Note: Duration tracking is not available atm due to restriction of the newrelic golang library. 