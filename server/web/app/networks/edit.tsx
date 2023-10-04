import { ChangeEvent, FormEvent, MouseEvent, useState } from 'react'
import Button from 'react-bootstrap/Button'
import Modal from 'react-bootstrap/Modal'
import Form from 'react-bootstrap/Form'
import { PutWithBody, Get, GetUrl } from '../api-client'
import { Network } from '../api-client/models'
import { useSWRConfig } from 'swr'
import { alertService } from '../services/alerts'
import { faEdit } from '@fortawesome/free-solid-svg-icons'
import { FontAwesomeIcon } from '@fortawesome/react-fontawesome'
import { OneOrBothBySize } from '../utils'

export default function Edit(props: { networkID: number }) {
    const { mutate } = useSWRConfig()
    const editNetwork = PutWithBody("/network")
    const networkEdited = Get("/network/" + props.networkID)
    const providers = Get("/providers")

    const networkCastObj: Network = {
        name: "",
    }
    const [network, setNetwork] = useState(networkCastObj)

    const [show, setShow] = useState(false)

    const handleClose = () => setShow(false)
    const handleShow = () => {
        setNetwork(networkEdited.data)
        setShow(true)
    }

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
        setNetwork(network => ({
            ...network,
            ...patchObject
        }));
    }

    async function submit(e: MouseEvent<HTMLButtonElement>) {
        const networkCast: Network = network
        // Currently name is the same as login
        networkCast.name = networkCast.login
        // @ts-ignore
        const data = await editNetwork.trigger(networkCast)
        if (data.status === 201) {
            let resJson = data.json()
            // Clean state
            setNetwork(networkCastObj)

            setValidated(false)
            handleClose()
            return mutate(GetUrl("/networks"))
        } else {
            alertService.error("Adding network failed with: ".concat(await data.text()), {})
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

    if (networkEdited.isLoading || providers.isLoading) return (<></>)

    return (
        <>
            <Button variant="primary" onClick={handleShow}>
            {OneOrBothBySize("Edit", <FontAwesomeIcon icon={faEdit} />)}
            </Button>
            <Modal show={show} onHide={handleClose}>
                <Form noValidate validated={validated} onSubmit={handleSubmit}>
                    <Modal.Header closeButton>
                        <Modal.Title>Add network</Modal.Title>
                    </Modal.Header>
                    <Modal.Body>
                        <Form.Group className="mb-3" controlId="formLogin">
                            <Form.Label>Login</Form.Label>
                            <Form.Control type="email" placeholder="Enter login" name="login" onChange={onChangeSetObject} defaultValue={network.login} required />
                            <Form.Text className="text-muted">
                                We will never share your email with anyone else.
                            </Form.Text>
                        </Form.Group>

                        <Form.Group className="mb-3" controlId="formPassword">
                            <Form.Label>Password</Form.Label>
                            <Form.Control type="password" placeholder="Password" name="password" onChange={onChangeSetObject} defaultValue={network.password} required />
                        </Form.Group>

                        <Form.Group className="mb-3" controlId="form2FA">
                            <Form.Label>2FA</Form.Label>
                            <Form.Control type="text" placeholder="2FA" name="twofa" onChange={onChangeSetObject} required />
                        </Form.Group>

                        <Form.Group className="mb-3" controlId="formNetworkType">
                            <Form.Label>Network type</Form.Label>
                            <Form.Select id="type" name="provider.name" onChange={onChangeSetObject} required>
                                {providers.data.map((item: any) => (
                                    <option key={item.ID}>{item.name}</option>
                                ))}
                            </Form.Select>
                        </Form.Group>
                    </Modal.Body>
                    <Modal.Footer>
                        <Button variant="secondary" onClick={handleClose}>
                            Close
                        </Button>
                        <Button variant="primary" type="submit" onClick={submit}>
                            Save Changes
                        </Button>
                    </Modal.Footer>
                </Form>
            </Modal>
        </>
    );
}

