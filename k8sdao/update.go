package k8sdao

import (
	"context"
	"fmt"

	"k8s.io/apimachinery/pkg/api/errors"

	v1 "k8s.io/api/apps/v1"
	"sigs.k8s.io/controller-runtime/pkg/log"

	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

func UpdateReplicas(ctx context.Context, replicas int32, nsedName types.NamespacedName, cli client.Client, ns string) error {
	deploy := v1.Deployment{}
	err := cli.Get(ctx, nsedName, &deploy)
	if err != nil {
		log.Log.Error(err, "get deploy error")
		return fmt.Errorf("%s", "get deploy error")
	}

	if errors.IsNotFound(err) {
		log.Log.Error(err, "deploy not found")
		return fmt.Errorf("%s", "deploy not found")
	} else if statusError, isStatus := err.(*errors.StatusError); isStatus {
		fmt.Printf("Error getting deployment%v\n", statusError.ErrStatus.Message)
	}

	name := deploy.GetName()
	log.Log.Info("%s deploy old replicas is %s: ", name, deploy.Spec.Replicas)

	deploy.Spec.Replicas = &replicas

	err = cli.Update(ctx, &deploy)
	if err != nil {
		return fmt.Errorf("%s", "update replicas error")
	}

	return nil
}
