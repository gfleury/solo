'use client'

import { usePathname } from 'next/navigation'
import Container from 'react-bootstrap/Container'
import Navbar from 'react-bootstrap/Navbar'
import Nav from 'react-bootstrap/Nav'
import Link from 'next/link'
import NavDropdown from 'react-bootstrap/NavDropdown'
import Image from './image'
import { User } from './api-client/models'
import NewButton from './newButton'
import { Stack } from 'react-bootstrap'
import { OneOrOtherBySize } from './utils'

const navigation = [
  { name: 'Home', href: '/' },
  { name: 'Networks', href: '/networks' },
  { name: 'Nodes', href: '/nodes' },
];

function signOut() {

}

function signIn(x: string) {

}

export default function OurNavbar({ user }: { user?: User }) {
  const pathname = usePathname();

  return (
    <Navbar collapseOnSelect expand="lg" className="bg-body-tertiary">
      <Container fluid>
        <Navbar.Brand>
          <Link
            href="/"
            className='followParent'>
            solo
          </Link>
        </Navbar.Brand>
        <Navbar.Toggle aria-controls="responsive-navbar-nav" />
        <Navbar.Collapse id="responsive-navbar-nav">
          <Nav className="me-auto">
            {navigation.map((item) => (
              <Link
                className={pathname === item.href ? 'nav-link active' : 'nav-link'}
                key={item.name}
                href={item.href}>
                {item.name}
              </Link>
            ))}
          </Nav>
          <Nav>
            <Stack direction={OneOrOtherBySize("horizontal", undefined)} gap={2}>
              <div className="p-2">
                <NewButton />
              </div>
              <div className="p-2">
                <NavDropdown title={(<Image
                  className="d-inline h-8 w-8 rounded-full"
                  src={user?.avatar || 'https://avatar.vercel.sh/leerob'}
                  height={25}
                  width={32}
                  alt={`${user?.email || 'placeholder'} avatar`}
                  referrerPolicy="no-referrer"
                />)} id="basic-nav-dropdown">
                  <NavDropdown.Header>Logged as {user?.email}</NavDropdown.Header>
                  <NavDropdown.Item href="#Logout">Logout</NavDropdown.Item>
                </NavDropdown>
              </div>
            </Stack>
          </Nav>
        </Navbar.Collapse>
      </Container>
    </Navbar>
  )
}
