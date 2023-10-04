'use client'

import { MouseEvent } from 'react'
import { Get, Delete, Put, GetUrl, HandleFailures } from '../api-client'
import { Account } from '../api-client/models'
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

export default function Accounts() {
    const { mutate } = useSWRConfig()
    const router = useRouter()
    const accounts = HandleFailures(Get("/accounts"), router)
    const accountDelete = Delete("/account/")
    const accountLogin = Put("/account/")

    function deleteFunction(accountID: number) {
        return async function deleteOnClick(e: MouseEvent<HTMLButtonElement>) {
            // @ts-ignore
            const data = await accountDelete.trigger(accountID)

            if (accountDelete.error || data.status != 204) {
                alertService.error("Deleting account failed with: ".concat(await data.text()), {})
            } else {
                mutate(GetUrl("/accounts"))
            }

        }
    }

    function loginFunction(accountID: number) {
        return async function loginOnClick(e: MouseEvent<HTMLButtonElement>) {
            // @ts-ignore
            const data = await accountLogin.trigger("" + accountID + "/session")
            if (accountDelete.error || data.status != 201) {
                alertService.error("Login account failed with: ".concat(await data.text()), {})
            } else {
                mutate(GetUrl("/accounts"))
            }

        }
    }

    function accountLoginButton(account: Account) {
        if (account.session == null) {
            return (
                <Button variant="primary" onClick={loginFunction(account.ID !== undefined ? account.ID : -1)}>
                    {OneOrBothBySize("Login", <FontAwesomeIcon icon={faRefresh} />)}
                </Button>)
        }
        return (<></>)
    }

    function accountCheckSession(account: Account) {
        if (account.provider.authenticationType == "oauth") {
            return (
                <Button variant="primary" onClick={loginFunction(account.ID !== undefined ? account.ID : -1)}>
                    {OneOrBothBySize("Refresh token", <FontAwesomeIcon icon={faRefresh} />)}
                </Button>)
        }
        return (<></>)
    }

    if (accounts.isLoading) return (Loading())

    const accountsCast: Account[] = accounts.data

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
                            <th>Account</th>
                            <th>Provider</th>
                            <th>Actions</th>
                        </tr>
                    </thead>
                    <tbody>
                        {accountsCast.map((account: Account) => (
                            <tr key={account.ID}>
                                <td>
                                    <Stack direction="horizontal" gap={2}>
                                    <div className="d-inline">
                                    <Image
                                        className="h-8 w-8 rounded-full"
                                        src={account?.avatar || 'https://upload.wikimedia.org/wikipedia/commons/a/ac/Default_pfp.jpg'}
                                        height={32}
                                        width={32}
                                        alt={`${account?.name || 'placeholder'} avatar`}
                                        referrerPolicy="no-referrer"
                                    />
                                    </div>
                                    <div className="d-inline">
                                    {account.name}
                                    </div>
                                    </Stack>
                                </td>
                                <td>{OneOrBothBySize(account.provider.name, <FontAwesomeIcon icon={getIcon(account.provider?.name)} />)}</td>
                                <td>
                                    <ButtonGroup size='sm'>
                                        <Button variant="primary" disabled={accountDelete.isMutating}
                                            onClick={deleteFunction(account.ID !== undefined ? account.ID : -1)}>
                                            {OneOrBothBySize("Delete", <FontAwesomeIcon icon={faRemove} />)}
                                        </Button>
                                        <Edit accountID={account.ID !== undefined ? account.ID : -1} />
                                        {accountLoginButton(account)}
                                        {accountCheckSession(account)}
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