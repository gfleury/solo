'use client'

import { ChangeEvent, FormEvent, MouseEvent, useState } from 'react'
import Button from 'react-bootstrap/Button'
import Form from 'react-bootstrap/Form'
import { useSWRConfig } from 'swr'
import { Post, GetUrl } from '../../api-client'
import { Network } from '../../api-client/models'
import { alertService } from '../../services/alerts'
import { useRouter } from 'next/navigation';
import { Stack } from 'react-bootstrap'

export default function Add() {
    const router = useRouter()
    const { mutate } = useSWRConfig()
    const addNetwork = Post("/network")

    const networkObj = {
        name: "",
        cidr: "",
        user: {},
        linkedusers: [],
    }
    const [network, setNetwork] = useState(networkObj)

    type FormControlElement = HTMLInputElement | HTMLTextAreaElement | HTMLSelectElement
    function onChangeSetObject(e: ChangeEvent<FormControlElement>) {
        const patchObject: any = {}
        const property = e.target.name
        patchObject[property] = e.target.value
        setNetwork(network => ({
            ...network,
            ...patchObject
        }));
    }

    async function submit(e: MouseEvent<HTMLElement>) {
        const networkCast: Network = network
        console.log(networkCast)
        // @ts-ignore
        const data = await addNetwork.trigger(networkCast)
        if (data.status === 201) {
            let resJson = data.json()
            // Clean state
            setNetwork(networkObj)

            setValidated(true)
            return router.push("/networks")
        } else {
            alertService.error("Adding network failed with: ".concat(await data.text()), {})
        }
    }

    const [validated, setValidated] = useState(false);

    function handleSubmit(event: FormEvent<HTMLFormElement>) {
        const form = event.currentTarget;
        event.preventDefault();
        event.stopPropagation();
    };

    return (
        <Stack className="my-2" gap={2}>
            <Form noValidate validated={validated} onSubmit={handleSubmit}>
                <Form.Label>Add network</Form.Label>
                <Form.Group>
                    <Form.Group className="mb-3" controlId="formName">
                        <Form.Label>Name</Form.Label>
                        <Form.Control type="name" placeholder="Network name, eg: Dummy Network" name="name" onChange={onChangeSetObject} required />
                        <Form.Text className="text-muted">
                            The network name, this is just a reference for you.
                        </Form.Text>
                    </Form.Group>

                    <Form.Group className="mb-3" controlId="formCIDR">
                        <Form.Label>Network CIDR</Form.Label>
                        <Form.Control type="networkaddress" placeholder="Network CIDR, eg: 10.1.0.0/24" name="cidr" onChange={onChangeSetObject} required />
                    </Form.Group>
                </Form.Group>
                <Form.Group>
                    <Button variant="secondary">
                        Cancel
                    </Button>&nbsp;
                    <Button variant="primary" type="submit" onClick={submit}>
                        Save Changes
                    </Button>
                </Form.Group>
            </Form>
        </Stack>
    );
}

