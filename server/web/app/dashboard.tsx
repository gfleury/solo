"use client";

import Container from "react-bootstrap/Container";
import Row from "react-bootstrap/Row";
import Nav from "./nav";
import { Suspense } from "react";
import MyAlert from "./alerts";
import { useState } from "react";
import { SessionContext, SessionStore } from "./context";

export default function Dashboard({ children }: { children: React.ReactNode }) {
  const sessionStore: SessionStore = {};
  const [session, setSession] = useState(sessionStore);

  return (
    <SessionContext.Provider value={{ session, setSession }}>
      <Container>
        <Row>
          <Suspense>
            <Nav />
          </Suspense>
        </Row>
        <Row>
          <Suspense>{children}</Suspense>
        </Row>
        <Row>
          <MyAlert />
        </Row>
      </Container>
    </SessionContext.Provider>
  );
}
