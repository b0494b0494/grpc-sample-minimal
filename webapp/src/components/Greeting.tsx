import React from 'react';
import { Form, Button } from 'react-bootstrap';
import { useGreeting } from '../hooks';

export const Greeting: React.FC = () => {
  const { name, setName, greeting, handleSayHello } = useGreeting();

  return (
    <section className="bg-light rounded p-4 border shadow-sm">
      <h2 className="h4 fw-semibold mb-3">Unary RPC: SayHello</h2>
      <Form onSubmit={handleSayHello} className="d-flex flex-column flex-sm-row gap-2 mb-3">
        <Form.Control
          type="text"
          value={name}
          onChange={(e) => setName(e.target.value)}
          placeholder="Enter your name"
          className="flex-grow-1"
        />
        <Button variant="primary" type="submit">
          Say Hello
        </Button>
      </Form>
      {greeting && (
        <div className={`p-3 rounded-md ${
          greeting.startsWith('Error') || greeting.startsWith('Network Error') 
            ? 'bg-red-50 text-red-700 border border-red-200' 
            : 'bg-green-50 text-green-700 border border-green-200'
        }`}>
          <p className="font-medium">{greeting}</p>
        </div>
      )}
    </section>
  );
};
