'use client'

import { ChangeEvent, FormEvent, MouseEvent, useState } from 'react'
import Button from 'react-bootstrap/Button'
import Modal from 'react-bootstrap/Modal'
import Form from 'react-bootstrap/Form'
import { useSWRConfig } from 'swr'
import { Post, Get, GetUrl } from '../../api-client'
import { Account } from '../../api-client/models'
import { alertService } from '../../services/alerts'
import { FormGroup, FormLabel } from 'react-bootstrap'

export default function Add() {
    const { mutate } = useSWRConfig()
    const addAccount = Post("/account")
    const providers = Get("/providers?type=loginpassword")

    const accountObj = {
        name: "",
        login: "",
        password: "",
        provider: { name: "Instagram" }
    }
    const [account, setAccount] = useState(accountObj)

    type FormControlElement = HTMLInputElement | HTMLTextAreaElement | HTMLSelectElement
    function onChangeSetObject(e: ChangeEvent<FormControlElement>) {
        const patchObject: any = {}
        const property = e.target.name
        if (property == "provider.name") {
            patchObject["provider"] = {
                name: e.target.value
            }
        } else {
            patchObject[property] = e.target.value
        }
        setAccount(account => ({
            ...account,
            ...patchObject
        }));
    }

    async function submit(e: MouseEvent<HTMLElement>) {
        const accountCast: Account = account
        // Currently name is the same as login
        accountCast.name = accountCast.login
        // @ts-ignore
        const data = await addAccount.trigger(accountCast)
        if (data.status === 201) {
            let resJson = data.json()
            // Clean state
            setAccount(accountObj)

            setValidated(false)
            return mutate(GetUrl("/accounts"))
        } else {
            alertService.error("Adding account failed with: ".concat(await data.text()), {})
        }

    }

    const [validated, setValidated] = useState(false);

    function handleSubmit(event: FormEvent<HTMLFormElement>) {
        const form = event.currentTarget;
        if (form.checkValidity() === false) {
            event.preventDefault();
            event.stopPropagation();
        }

        setValidated(true);
    };

    if (providers.isLoading) return (<></>)

    return (
        <>
            <Form noValidate validated={validated} onSubmit={handleSubmit}>
                <Form.Label closeButton>Add account</Form.Label>
                <Form.Group>
                    <Form.Group className="mb-3" controlId="formLogin">
                        <Form.Label>Login</Form.Label>
                        <Form.Control type="email" placeholder="Enter login" name="login" onChange={onChangeSetObject} required />
                        <Form.Text className="text-muted">
                            We will never share your email with anyone else.
                        </Form.Text>
                    </Form.Group>

                    <Form.Group className="mb-3" controlId="formPassword">
                        <Form.Label>Password</Form.Label>
                        <Form.Control type="password" placeholder="Password" name="password" onChange={onChangeSetObject} required />
                    </Form.Group>

                    <Form.Group className="mb-3" controlId="form2FA">
                        <Form.Label>2FA</Form.Label>
                        <Form.Control type="text" placeholder="2FA" name="twofa" onChange={onChangeSetObject} required />
                    </Form.Group>

                    <Form.Group className="mb-3" controlId="formAccountType">
                        <Form.Label>Account type</Form.Label>
                        <Form.Select id="type" name="provider.name" onChange={onChangeSetObject} required>
                            {providers.data.map((item: any) => (
                                <option key={item.ID}>{item.name}</option>
                            ))}
                        </Form.Select>
                    </Form.Group>
                </Form.Group>
                <Form.Group>
                    <Button variant="secondary">
                        Cancel
                    </Button>
                    <Button variant="primary" type="submit" onClick={submit}>
                        Save Changes
                    </Button>
                </Form.Group>
            </Form>
        </>
    );
}

