'use client'
import { useContext } from 'react'
import { SessionContext } from '../context'
import { Stack } from 'react-bootstrap'

export default function Home() {
    const { session, setSession } = useContext(SessionContext)

    return (
        <Stack gap={2}>
            <div className="p-2">
                Hellow {session?.user?.email}!
            </div>
        </Stack>
    )
}