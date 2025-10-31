import React from 'react';
import { Button } from 'react-bootstrap';
import { useCounterStream } from '../hooks';

export const CounterStream: React.FC = () => {
  const { counterOutput, handleStartCounterStream } = useCounterStream();

  return (
    <section className="bg-light rounded p-4 border shadow-sm">
      <h2 className="h4 fw-semibold mb-3">Server-Side Streaming RPC: StreamCounter</h2>
      <Button 
        variant="primary"
        onClick={handleStartCounterStream}
        className="mb-3"
      >
        Start Counter Stream
      </Button>
      <div className="p-3 border bg-white rounded" style={{ minHeight: '100px' }}>
        {counterOutput.length === 0 ? (
          <p className="text-muted small">Click the button above to start the counter stream.</p>
        ) : (
          counterOutput.map((item, index) => (
            <p key={index} className="py-2 border-bottom border-light small font-monospace">{item}</p>
          ))
        )}
      </div>
    </section>
  );
};
