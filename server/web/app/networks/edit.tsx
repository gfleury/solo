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

    const networkCastObj: Network = {
        name: "",
        cidr: "",
        connection_config: "",
        user: {},
        linkedusers: [],
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
        patchObject[property] = e.target.value
        setNetwork(network => ({
            ...network,
            ...patchObject
        }));
    }

    async function submit(e: MouseEvent<HTMLButtonElement>) {
        const networkCast: Network = network
        try {
            // @ts-ignore
            const data = await editNetwork.trigger(networkCast)
            if (data.status === 201) {
                let resJson = data.json()
                // Clean state
                setNetwork(networkCastObj)

                setValidated(true)
                handleClose()
                setValidated(false)

                return mutate(GetUrl("/networks"))
            } else {
                alertService.error("Adding network failed with: ".concat(await data.text()), {})
            }
        } catch (error) {
            console.log(error)
        }
    }

    const [validated, setValidated] = useState(false);

    function handleSubmit(event: FormEvent<HTMLFormElement>) {
        const form = event.currentTarget;
        event.preventDefault();
        event.stopPropagation();
    };

    if (networkEdited.isLoading) return (<></>)

    return (
        <>
            <Button variant="primary" onClick={handleShow}>
                {OneOrBothBySize("Edit", <FontAwesomeIcon icon={faEdit} />)}
            </Button>
            <Modal show={show} onHide={handleClose}>
                <Form noValidate validated={validated} onSubmit={handleSubmit}>
                    <Modal.Header closeButton>
                        <Modal.Title>Edit network</Modal.Title>
                    </Modal.Header>
                    <Modal.Body>
                        <Form.Group>
                            <Form.Group className="mb-3" controlId="formName">
                                <Form.Label>Name</Form.Label>
                                <Form.Control type="name" placeholder="Network name, eg: Dummy Network" name="name" onChange={onChangeSetObject} defaultValue={network.name} required />
                                <Form.Text className="text-muted">
                                    The network name, this is just a reference for you.
                                </Form.Text>
                            </Form.Group>

                            <Form.Group className="mb-3" controlId="formCIDR">
                                <Form.Label>Network CIDR</Form.Label>
                                <Form.Control type="networkaddress" placeholder="Network CIDR, eg: 10.1.0.0/24" name="cidr" onChange={onChangeSetObject} defaultValue={network.cidr} required />
                            </Form.Group>

                            <Form.Group className="mb-3" controlId="formConnectionConfiguration">
                                <Form.Label>Network CIDR</Form.Label>
                                <Form.Control as="textarea" rows={3} placeholder="This must be the connection configuration token" name="connection_config" onChange={onChangeSetObject} defaultValue={network.connection_config} required />
                            </Form.Group>
                        </Form.Group>
                    </Modal.Body>
                    <Form.Group>
                        <Button variant="secondary" onClick={handleClose}>
                            Cancel
                        </Button>&nbsp;
                        <Button variant="primary" type="submit" onClick={submit}>
                            Save Changes
                        </Button>
                    </Form.Group>
                </Form>
            </Modal>
        </>
    );
}

