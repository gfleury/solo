'use client'

import { ChangeEvent, FormEvent, MouseEvent, useState } from 'react'
import Button from 'react-bootstrap/Button'
import Form from 'react-bootstrap/Form'
import { Put, Get } from '../../api-client'
import { Network } from '../../api-client/models'
import { alertService } from '../../services/alerts'
import { useRouter } from 'next/navigation';

export default function Add() {
    const router = useRouter()
    const registerNode = Put("/network")
    const networks = Get("/networks")

    const activationObj = {
        code: "",
        networkID: "0",
    }
    const [activation, setActivation] = useState(activationObj)

    type FormControlElement = HTMLInputElement | HTMLTextAreaElement | HTMLSelectElement
    function onChangeSetObject(e: ChangeEvent<FormControlElement>) {
        const patchObject: any = {}
        const property = e.target.name
        patchObject[property] = e.target.value
        setActivation(activation => ({
            ...activation,
            ...patchObject
        }));
        console.log(activation)
    }

    async function submit(e: MouseEvent<HTMLElement>) {
        console.log(activation)
        // @ts-ignore
        const data = await registerNode.trigger(`/${activation.networkID}/${activation.code}`)
        if (data.status === 200) {
            let resJson = data.json()
            // Clean state
            setActivation(activationObj)

            setValidated(true)
            return router.push("/networks")
        } else {
            alertService.error("Adding node failed with: ".concat(await data.text()), {})
        }
    }

    const [validated, setValidated] = useState(false);

    function handleSubmit(event: FormEvent<HTMLFormElement>) {
        const form = event.currentTarget;
        event.preventDefault();
        event.stopPropagation();
    };

    if (networks.isLoading) return (<></>)
    
    return (
        <>
            <Form noValidate validated={validated} onSubmit={handleSubmit}>
                <Form.Label>Register Node</Form.Label>
                <Form.Group>
                    <Form.Group className="mb-3" controlId="formName">
                        <Form.Label>Activation code</Form.Label>
                        <Form.Control type="name" placeholder="Activation code from cli" name="code" onChange={onChangeSetObject} required />
                        <Form.Text className="text-muted">
                            The code you got when you ran `$ solo register`.
                        </Form.Text>
                    </Form.Group>

                    <Form.Group className="mb-3" controlId="formNetwork">
                        <Form.Label>Network</Form.Label>
                        <Form.Select id="type" name="networkID" onChange={onChangeSetObject} required>
                            <option key="0" value="0">Choose a network to register the node</option>
                            {networks.data.map((item: any) => (
                                <option key={item.ID} value={item.ID}>{item.name}</option>
                            ))}
                        </Form.Select>
                    </Form.Group>
                </Form.Group>
                <Form.Group>
                    <Button variant="secondary">
                        Cancel
                    </Button>&nbsp;
                    <Button variant="primary" type="submit" onClick={submit}>
                        Register
                    </Button>
                </Form.Group>
            </Form>
        </>
    );
}

