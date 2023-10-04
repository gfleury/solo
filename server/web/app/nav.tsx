'use client'
import { useContext, useEffect } from 'react';
import { SessionContext } from './context';
import { ForceLogin, Get } from './api-client'
import Loading from './loading'
import Navbar from './navbar';
import { Button } from 'react-bootstrap';
import { useRouter } from 'next/navigation';
import { FontAwesomeIcon } from '@fortawesome/react-fontawesome'
import { faGoogle } from '@fortawesome/free-brands-svg-icons'

export default function Nav() {
  const { session, setSession } = useContext(SessionContext)

  const { data, isLoading, error } = Get("/user/login")

  const router = useRouter()

  useEffect(() => {
    if (setSession) setSession({ user: data })
  }, [setSession, data])

  if (!setSession || !session) { return (<>Failed to load session</>) }
  if (isLoading) return Loading()
  if (error) return <Button onClick={() => router.push(ForceLogin())}><FontAwesomeIcon icon={faGoogle} /> Login with Google</Button>

  return <Navbar user={session.user} />;
}
