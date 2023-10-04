'use client'

import { MouseEvent } from 'react'
import { Get, Delete, Put, GetUrl, HandleFailures } from '../api-client'
import { Network } from '../api-client/models'
import Loading from '../loading'
import Table from 'react-bootstrap/Table';
import Edit from './edit'
import Button from 'react-bootstrap/Button'
import ButtonGroup from 'react-bootstrap/ButtonGroup'
import { useSWRConfig } from 'swr'
import { alertService } from '../services/alerts'
import Image from '../image'
import { faDev, faInstagram } from '@fortawesome/free-brands-svg-icons'
import { FontAwesomeIcon } from '@fortawesome/react-fontawesome'
import SearchBar from './search'
import { Stack } from 'react-bootstrap'
import { faRefresh, faRemove } from '@fortawesome/free-solid-svg-icons'
import { OneOrBothBySize } from '../utils'
import { useRouter } from 'next/navigation'

export default function Networks() {
    const { mutate } = useSWRConfig()
    const router = useRouter()
    const networks = HandleFailures(Get("/networks"), router)
    const networkDelete = Delete("/network/")
    const networkLogin = Put("/network/")

    function deleteFunction(networkID: number) {
        return async function deleteOnClick(e: MouseEvent<HTMLButtonElement>) {
            // @ts-ignore
            const data = await networkDelete.trigger(networkID)

            if (networkDelete.error || data.status != 204) {
                alertService.error("Deleting network failed with: ".concat(await data.text()), {})
            } else {
                mutate(GetUrl("/networks"))
            }

        }
    }

    function loginFunction(networkID: number) {
        return async function loginOnClick(e: MouseEvent<HTMLButtonElement>) {
            // @ts-ignore
            const data = await networkLogin.trigger("" + networkID + "/session")
            if (networkDelete.error || data.status != 201) {
                alertService.error("Login network failed with: ".concat(await data.text()), {})
            } else {
                mutate(GetUrl("/networks"))
            }

        }
    }

    function networkLoginButton(network: Network) {
        if (network.session == null) {
            return (
                <Button variant="primary" onClick={loginFunction(network.ID !== undefined ? network.ID : -1)}>
                    {OneOrBothBySize("Login", <FontAwesomeIcon icon={faRefresh} />)}
                </Button>)
        }
        return (<></>)
    }

    function networkCheckSession(network: Network) {
        if (network.provider.authenticationType == "oauth") {
            return (
                <Button variant="primary" onClick={loginFunction(network.ID !== undefined ? network.ID : -1)}>
                    {OneOrBothBySize("Refresh token", <FontAwesomeIcon icon={faRefresh} />)}
                </Button>)
        }
        return (<></>)
    }

    if (networks.isLoading) return (Loading())

    const networksCast: Network[] = networks.data

    function getIcon(name: string | undefined) {
        if (name === "Instagram") {
            return faInstagram
        }
        return faDev
    }

    return (
        <Stack className="my-2" gap={2}>
            <div>
                <SearchBar />
            </div>
            <div>
                <Table align="center" responsive>
                    <thead>
                        <tr>
                            <th>Network</th>
                            <th>Provider</th>
                            <th>Actions</th>
                        </tr>
                    </thead>
                    <tbody>
                        {networksCast.map((network: Network) => (
                            <tr key={network.ID}>
                                <td>
                                    <Stack direction="horizontal" gap={2}>
                                    <div className="d-inline">
                                    <Image
                                        className="h-8 w-8 rounded-full"
                                        src={network?.avatar || 'https://upload.wikimedia.org/wikipedia/commons/a/ac/Default_pfp.jpg'}
                                        height={32}
                                        width={32}
                                        alt={`${network?.name || 'placeholder'} avatar`}
                                        referrerPolicy="no-referrer"
                                    />
                                    </div>
                                    <div className="d-inline">
                                    {network.name}
                                    </div>
                                    </Stack>
                                </td>
                                <td>{OneOrBothBySize(network.provider.name, <FontAwesomeIcon icon={getIcon(network.provider?.name)} />)}</td>
                                <td>
                                    <ButtonGroup size='sm'>
                                        <Button variant="primary" disabled={networkDelete.isMutating}
                                            onClick={deleteFunction(network.ID !== undefined ? network.ID : -1)}>
                                            {OneOrBothBySize("Delete", <FontAwesomeIcon icon={faRemove} />)}
                                        </Button>
                                        <Edit networkID={network.ID !== undefined ? network.ID : -1} />
                                        {networkLoginButton(network)}
                                        {networkCheckSession(network)}
                                    </ButtonGroup>
                                </td>
                            </tr>
                        ))}
                    </tbody>
                </Table>
            </div>
        </Stack>
    )
}